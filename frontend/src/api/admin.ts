import { apiGet } from "./client";
import type { AdminDashboardSummary } from "../types/admin";
import type { MenuResponse } from "../types/menu";
import { getAdminTables } from "./tables";

export async function getAdminDashboardSummary(): Promise<AdminDashboardSummary> {
  const [menu, tables] = await Promise.all([apiGet<MenuResponse>("/api/v1/menu"), getAdminTables()]);

  const menuItems = menu.items.length;
  const categories = menu.categories.length;
  const availableItems = menu.items.filter((item) => item.is_available).length;
  const soldOutItems = menuItems - availableItems;
  const featuredItems = menu.items.filter((item) => item.is_featured).length;

  return {
    menuItems,
    categories,
    tables: tables.length,
    availableItems,
    soldOutItems,
    featuredItems,
  };
}
