import { apiGet } from "./client";
import type { MenuResponse } from "../types/menu";

export function getMenu(): Promise<MenuResponse> {
  return apiGet<MenuResponse>("/api/v1/menu");
}
