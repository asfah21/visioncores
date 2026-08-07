# GAVIS (GSI AI Vision for Intelligent Surveillance)

**GAVIS** is an internal innovation initiative developed at Site Wolo, designed as a strategic alternative and benchmark to third-party AI-powered CCTV solutions. This project reflects our commitment to building a robust, intelligent, and fully controlled surveillance system tailored to the operational needs of our organization.

Unlike conventional vendor-based systems, GAVIS is developed internally, providing greater flexibility and adaptability. The platform is engineered to align with the specific conditions of our operational environments, allowing us to customize features, detection models, and system behavior based on real field requirements.

## Key Advantages

- **Data Security**: By maintaining full ownership of the infrastructure and data processing pipeline, GAVIS eliminates dependency on third-party vendors, reducing exposure to external risks.
- **Continuous Improvement**: We train and deploy our own AI models. This ensures that the system becomes increasingly accurate and context-aware over time, particularly in recognizing patterns, personnel, and activities unique to our sites.
- **Long-term Efficiency**: Provides scalability, cost control, and seamless system integration without vendor limitations.
- **Live Monitoring & Analytics**: Features live CCTV streams with AI bounding boxes, presence heatmaps, and peak traffic analytics.

## Tech Stack

This project is built using modern web development standards and a responsive design system:

- **Framework**: Next.js (App Router), TypeScript
- **Styling**: Tailwind CSS v4
- **UI Components**: Shadcn UI & Lucide Icons
- **Real-time Data**: WebSocket for live AI detections and CCTV feeds
- **Tooling**: Biome, Husky

## Getting Started

1. **Clone the repository**
   ```bash
   git clone <internal-repo-url>
   ```
   
2. **Navigate into the project**
   ```bash
   cd frontend
   ```
   
3. **Install dependencies**
   ```bash
   npm install
   ```

4. **Start the development server**
   ```bash
   npm run dev
   ```

Your app will be running at [http://localhost:3000](http://localhost:3000).

> **Note:** You must be connected to the internal network (or VPN) to receive data correctly from the Live CCTV server and WebSocket APIs (e.g., `http://10.10.11.5`).

## Formatting and Linting

Format, lint, and organize imports:
```bash
npx @biomejs/biome check --write
```

---

We present this initiative as a forward-looking solution that prioritizes flexibility, security, and sustainability in modern surveillance systems.
