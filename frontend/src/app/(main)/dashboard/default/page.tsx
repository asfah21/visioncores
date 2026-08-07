"use client";

import { useState } from "react";

import { InsightCards } from "./_components/insight-cards";
import { LiveCctvCard } from "./_components/live-cctv-card";
import { OverviewCards } from "./_components/overview-cards";

export default function Page() {
  const [selectedCam, setSelectedCam] = useState("cam1");

  return (
    <div className="flex flex-col gap-4 p-0">
      <div className="grid grid-cols-1 items-stretch gap-4 xl:grid-cols-2">
        <LiveCctvCard className="h-full" selectedCam={selectedCam} onCamChange={setSelectedCam} />
        <div className="h-full">
          <OverviewCards className="h-full gap-4 xl:grid-cols-2" cameraId={selectedCam} />
        </div>
      </div>
      <div className="grid grid-cols-1 gap-4">
        <InsightCards />
        {/* <OperationalCards /> */}
        {/* <RecentLeadsTable data={recentLeadsData} /> */}
      </div>
    </div>
  );
}
