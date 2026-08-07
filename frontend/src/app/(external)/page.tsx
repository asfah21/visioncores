import { redirect } from "next/navigation";

export default function Home() {
  redirect("/auth/v1/login");
}
