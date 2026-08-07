# main.py - YOLO Original Optimized (Stabil & Terbukti) + Behavior Detection
import cv2
import requests
import time
import os
import numpy as np
from insightface.app import FaceAnalysis
from ultralytics import YOLO

# ================= CONFIG =================
API_URL = "http://backend:3001/detection"

# ✅ NEW: support env (fallback ke default lama)
CAMERA_ID = os.getenv("CAMERA_ID", "cam1")
RTSP_URL = os.getenv("RTSP_URL", "")
if not RTSP_URL:
    raise RuntimeError("RTSP_URL must be set for the detector worker")

CONF_THRESHOLD = 0.4
FRAME_SKIP = 2
KEEP_ALIVE = 10
MISSING_FRAMES = 5

FACE_SIM_THRESHOLD = 0.35
FACE_DETECT_SKIP = 10
RESIZE_W, RESIZE_H = 640, 360
SEND_INTERVAL = 0.5

# === BEHAVIOR CONFIG ===
DIRECTION_THRESHOLD = 5
LOITERING_TIME = 10  # detik
ZIGZAG_THRESHOLD = 3  # jumlah perubahan arah

# === POSTURE / ACTIVITY BEHAVIOR CONFIG ===
# Rasio H/W bounding box: orang berdiri = tinggi (rasio > threshold), duduk = lebih pendek
STANDING_RATIO_THRESHOLD = 1.8   # H/W > nilai ini = standing
SITTING_RATIO_THRESHOLD  = 1.1   # H/W antara ini dan STANDING = sitting; di bawah = lebih ambigu
# Pergerakan centroid untuk menentukan walking vs standing
WALKING_MOVEMENT_THRESHOLD = 8   # pixel per frame (setelah FRAME_SKIP), jika bergerak = walking
# Smoothing: pakai buffer N observasi sebelum kirim event posture
POSTURE_BUFFER_SIZE = 5
# Interval kirim event posture (detik), untuk menghindari spam
POSTURE_SEND_INTERVAL = 3.0

os.environ["OPENCV_FFMPEG_CAPTURE_OPTIONS"] = "rtsp_transport;tcp|max_delay;500000"

# ================= INITIALIZATION =================
print("[INIT] YOLO with tracking...")
model = YOLO("yolov8n.pt")

print("[INIT] FACE recognition...")
face_app = FaceAnalysis(name="buffalo_l")
face_app.prepare(ctx_id=-1)

# ================= FACE DATABASE =================
known_embeddings = []
known_names = []

def load_faces():
    path = "faces"
    if not os.path.exists(path):
        print("[FACE] Folder 'faces' tidak ditemukan")
        return
    for file in os.listdir(path):
        if not file.endswith(('.jpg', '.jpeg', '.png')):
            continue
        img = cv2.imread(f"{path}/{file}")
        if img is None:
            continue
        faces = face_app.get(img)
        if len(faces) > 0:
            known_embeddings.append(faces[0].embedding)
            known_names.append(os.path.splitext(file)[0])
            print(f"[FACE] Loaded: {file}")

def recognize_faces(frame):
    if not known_embeddings:
        return []
    
    faces = face_app.get(frame)
    results = []
    for face in faces:
        emb = face.embedding
        best_match = "Unknown"
        best_score = 0
        for known_emb, name in zip(known_embeddings, known_names):
            sim = np.dot(emb, known_emb) / (np.linalg.norm(emb) * np.linalg.norm(known_emb))
            if sim > best_score:
                best_score = sim
                best_match = name
        if best_score < FACE_SIM_THRESHOLD:
            best_match = "Unknown"
        x1, y1, x2, y2 = map(int, face.bbox)
        results.append((x1, y1, x2, y2, best_match))
    return results

def send_event(data):
    try:
        requests.post(API_URL, json=data, timeout=0.2)
    except Exception as e:
        print(f"[SEND ERROR] {e}")

def open_stream(url):
    cap = cv2.VideoCapture(url, cv2.CAP_FFMPEG)
    cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
    if cap.isOpened():
        print("[OK] RTSP CONNECTED")
        return cap
    print("[FAIL] RTSP FAILED")
    return None

# =====================================================
# 🧍 POSTURE DETECTION: walking / standing / sitting
# =====================================================
def classify_posture(bbox_w, bbox_h, movement_px):
    """
    Klasifikasi postur berdasarkan:
    - Rasio H/W bounding box  → standing vs sitting
    - Pergerakan centroid      → walking vs diam
    
    Return: "walking" | "standing" | "sitting"
    """
    ratio = bbox_h / max(bbox_w, 1)  # hindari division by zero

    if ratio >= STANDING_RATIO_THRESHOLD:
        # Orang berdiri tegak (bbox tinggi dan sempit)
        if movement_px >= WALKING_MOVEMENT_THRESHOLD:
            return "walking"
        else:
            return "standing"
    else:
        # Bbox lebih pendek/lebar → kemungkinan duduk
        return "sitting"


def get_smoothed_posture(posture_buffer):
    """
    Ambil postur mayoritas dari buffer untuk mengurangi fluktuasi.
    Return postur yang paling sering muncul dalam buffer.
    """
    if not posture_buffer:
        return None
    return max(set(posture_buffer), key=posture_buffer.count)

# ================= MAIN LOOP =================
def main_loop():
    print("[START] YOLO Tracking with Face Recognition")
    print("=" * 50)
    
    cap = None
    load_faces()
    
    frame_count = 0
    last_send = {}
    active_ids = set()
    last_seen = {}
    missing_count = {}
    track_to_name = {}
    track_face_lock = {}
    face_results = []
    error_count = 0
    reconnect_count = 0

    # === BEHAVIOR STATE ===
    track_history = {}       # posisi sebelumnya
    direction_history = {}   # arah sebelumnya
    direction_changes = {}   # hitung zigzag
    first_seen_time = {}     # untuk loitering

    # === POSTURE STATE ===
    posture_buffer = {}        # {track_id: [list of recent posture strings]}
    last_posture = {}          # {track_id: posture_string terakhir yang sudah dikirim}
    last_posture_send = {}     # {track_id: timestamp terakhir kirim posture event}
    
    while True:
        try:
            if cap is None or not cap.isOpened():
                reconnect_count += 1
                if reconnect_count > 3:
                    print(f"[RECONNECT] Attempt {reconnect_count}...")
                    time.sleep(2)
                cap = open_stream(RTSP_URL)
                if cap is None:
                    continue
                reconnect_count = 0
            
            for _ in range(2):
                cap.grab()
            
            ret, frame = cap.read()
            if not ret or frame is None:
                error_count += 1
                if error_count > 10:
                    print("[ERROR] Too many frame errors, reconnecting...")
                    cap.release()
                    cap = None
                    error_count = 0
                continue
            error_count = 0
            
            frame_count += 1
            if frame_count % FRAME_SKIP != 0:
                continue
            
            frame = cv2.resize(frame, (RESIZE_W, RESIZE_H))
            now = time.time()
            
            if frame_count % FACE_DETECT_SKIP == 0 and known_embeddings:
                face_results = recognize_faces(frame)
            
            results = model.track(
                frame,
                persist=True,
                conf=0.3,
                iou=0.7,
                tracker="botsort.yaml",
                verbose=False
            )
            
            current_ids = set()
            batch_detections = []
            
            for r in results:
                if r.boxes is None or r.boxes.id is None:
                    continue
                
                for box in r.boxes:
                    cls = int(box.cls)
                    if model.names[cls] != "person":
                        continue
                    
                    conf = float(box.conf)
                    if conf < CONF_THRESHOLD:
                        continue
                    
                    track_id = int(box.id)
                    x1, y1, x2, y2 = map(int, box.xyxy[0])

                    # === CENTROID ===
                    cx = int((x1 + x2) / 2)
                    cy = int((y1 + y2) / 2)
                    
                    area = (x2 - x1) * (y2 - y1)
                    if area < 4000:
                        continue
                    
                    current_ids.add(track_id)
                    last_seen[track_id] = now
                    missing_count[track_id] = 0

                    # === INIT FIRST SEEN ===
                    if track_id not in first_seen_time:
                        first_seen_time[track_id] = now
                    
                    # =====================
                    # 🚀 DIRECTION TRACKING
                    # =====================
                    prev = track_history.get(track_id)
                    movement_px = 0  # default jika belum ada history

                    if prev:
                        dx = cx - prev[0]
                        dy = cy - prev[1]
                        movement_px = (dx**2 + dy**2) ** 0.5  # jarak Euclidean centroid

                        if abs(dx) > DIRECTION_THRESHOLD:
                            curr_dir = "RIGHT" if dx > 0 else "LEFT"
                            prev_dir = direction_history.get(track_id)

                            if prev_dir and curr_dir != prev_dir:
                                # 🔁 TURN BACK EVENT
                                send_event({
                                    "camera_id": CAMERA_ID,
                                    "event": "turn_back",
                                    "person_id": track_id,
                                    "timestamp": int(now)
                                })

                                # count zigzag
                                direction_changes[track_id] = direction_changes.get(track_id, 0) + 1

                            direction_history[track_id] = curr_dir

                    track_history[track_id] = (cx, cy)

                    # =====================
                    # 🚨 SUSPICIOUS DETECTION
                    # =====================

                    # 🧍 LOITERING
                    if now - first_seen_time.get(track_id, now) > LOITERING_TIME:
                        send_event({
                            "camera_id": CAMERA_ID,
                            "event": "loitering",
                            "person_id": track_id,
                            "duration": int(now - first_seen_time[track_id]),
                            "timestamp": int(now)
                        })
                        first_seen_time[track_id] = now  # reset biar gak spam

                    # 🔁 ZIGZAG
                    if direction_changes.get(track_id, 0) >= ZIGZAG_THRESHOLD:
                        send_event({
                            "camera_id": CAMERA_ID,
                            "event": "zigzag_movement",
                            "person_id": track_id,
                            "count": direction_changes[track_id],
                            "timestamp": int(now)
                        })
                        direction_changes[track_id] = 0

                    # =====================
                    # 🏃 POSTURE DETECTION
                    # =====================
                    bbox_w = x2 - x1
                    bbox_h = y2 - y1

                    raw_posture = classify_posture(bbox_w, bbox_h, movement_px)

                    # Masukkan ke buffer smoothing per track_id
                    if track_id not in posture_buffer:
                        posture_buffer[track_id] = []
                    posture_buffer[track_id].append(raw_posture)
                    if len(posture_buffer[track_id]) > POSTURE_BUFFER_SIZE:
                        posture_buffer[track_id].pop(0)

                    smoothed_posture = get_smoothed_posture(posture_buffer[track_id])

                    # Kirim event hanya jika postur berubah DAN interval cukup
                    prev_posture = last_posture.get(track_id)
                    time_since_posture_send = now - last_posture_send.get(track_id, 0)

                    if (
                        smoothed_posture is not None
                        and smoothed_posture != prev_posture
                        and time_since_posture_send >= POSTURE_SEND_INTERVAL
                    ):
                        send_event({
                            "camera_id": CAMERA_ID,
                            "event": "behavior_change",
                            "behavior": smoothed_posture,  # "walking" | "standing" | "sitting"
                            "person_id": track_id,
                            "name": track_to_name.get(track_id, "Unknown"),
                            "timestamp": int(now)
                        })
                        last_posture[track_id] = smoothed_posture
                        last_posture_send[track_id] = now

                    # =====================
                    # FACE MATCH
                    # =====================
                    if track_id not in track_face_lock and face_results:
                        best_overlap = 0
                        best_name = None
                        for fx1, fy1, fx2, fy2, name in face_results:
                            overlap_w = min(x2, fx2) - max(x1, fx1)
                            overlap_h = min(y2, fy2) - max(y1, fy1)
                            if overlap_w > 0 and overlap_h > 0:
                                overlap_area = overlap_w * overlap_h
                                face_area = (fx2 - fx1) * (fy2 - fy1)
                                overlap_ratio = overlap_area / face_area if face_area > 0 else 0
                                if overlap_ratio > 0.3 and name != "Unknown":
                                    if overlap_ratio > best_overlap:
                                        best_overlap = overlap_ratio
                                        best_name = name
                        if best_name:
                            track_to_name[track_id] = best_name
                            track_face_lock[track_id] = True
                    
                    person_name = track_to_name.get(track_id, "Unknown")
                    
                    if now - last_send.get(track_id, 0) > SEND_INTERVAL:
                        batch_detections.append({
                            "camera_id": CAMERA_ID,
                            "person_id": track_id,
                            "name": person_name,
                            "confidence": round(conf, 2),
                            "bbox": {"x1": int(x1), "y1": int(y1), "x2": int(x2), "y2": int(y2)},
                            "frame_size": {"w": RESIZE_W, "h": RESIZE_H},
                            "posture": smoothed_posture,  # disertakan juga di batch detection
                            "timestamp": int(now)
                        })
                        last_send[track_id] = now
                    
                    if track_id not in active_ids:
                        send_event({
                            "camera_id": CAMERA_ID,
                            "event": "person_enter",
                            "person_id": track_id,
                            "name": person_name,
                            "timestamp": int(now)
                        })
                        active_ids.add(track_id)
            
            if batch_detections:
                send_event({
                    "event": "person_detect_batch",
                    "detections": batch_detections,
                    "total": len(batch_detections),
                    "timestamp": int(now)
                })
            
            for old_id in list(active_ids):
                if old_id not in current_ids:
                    missing_count[old_id] = missing_count.get(old_id, 0) + 1
                    if missing_count[old_id] >= MISSING_FRAMES:
                        if now - last_seen.get(old_id, 0) > KEEP_ALIVE:
                            send_event({
                                "camera_id": CAMERA_ID,
                                "event": "person_exit",
                                "person_id": old_id,
                                "name": track_to_name.get(old_id, "Unknown"),
                                "timestamp": int(now)
                            })
                            track_to_name.pop(old_id, None)
                            track_face_lock.pop(old_id, None)
                            last_seen.pop(old_id, None)
                            missing_count.pop(old_id, None)
                            # Cleanup posture state untuk ID yang sudah exit
                            posture_buffer.pop(old_id, None)
                            last_posture.pop(old_id, None)
                            last_posture_send.pop(old_id, None)
                            active_ids.discard(old_id)
            
            if frame_count % 100 == 0:
                print(f"[STATUS] Frame: {frame_count}, Active: {len(active_ids)}")
                
        except Exception as e:
            print(f"[ERROR] {e}")
            time.sleep(1)
    
    if cap:
        cap.release()

if __name__ == "__main__":
    main_loop()
