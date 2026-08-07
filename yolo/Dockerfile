FROM python:3.10

WORKDIR /app

RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    libgomp1 \
    wget \
    && rm -rf /var/lib/apt/lists/*

# 🔥 Install OpenVINO
RUN pip install --no-cache-dir \
    openvino==2024.6.0 \
    ultralytics==8.3.133 \
    opencv-python-headless \
    requests \
    insightface==0.7.3 \
    onnxruntime==1.18.0 \
    numpy==1.24.3

COPY . .

CMD ["python", "main.py"]
