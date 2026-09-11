import { useEffect } from "react";
import { Route, Routes, useParams } from "react-router-dom";
import { StickyMobileCartBar } from "./components/cart/StickyMobileCartBar";
import { CartProvider, useCart } from "./context/CartContext";
import { getAdminTables } from "./api/tables";
import CartPage from "./pages/CartPage";
import AdminCategoriesPage from "./pages/AdminCategoriesPage";
import AdminDashboardPage from "./pages/AdminDashboardPage";
import AdminMenuPage from "./pages/AdminMenuPage";
import AdminGalleryPage from "./pages/AdminGalleryPage";
import AdminSettingsPage from "./pages/AdminSettingsPage";
import AdminReviewsPage from "./pages/AdminReviewsPage";
import AdminBannersPage from "./pages/AdminBannersPage";
import AdminTablesPage from "./pages/AdminTablesPage";
import AdminKitchenAvailabilityPage from "./pages/AdminKitchenAvailabilityPage";
import AdminLoginPage from "./pages/AdminLoginPage";
import HomePage from "./pages/HomePage";
import MenuPage from "./pages/MenuPage";
import NotFoundPage from "./pages/NotFoundPage";
import { AdminRoute } from "./components/common/AdminRoute";
import { NetworkStatus } from "./components/common/NetworkStatus";

function TableRouteMenuPage() {
  const { tableNumber, token } = useParams();
  const { setTableInfo } = useCart();

  useEffect(() => {
    if (!tableNumber && !token) {
      return;
    }

    let cancelled = false;

    void getAdminTables().then((tables) => {
      if (cancelled) {
        return;
      }

      const match = tables.find((table) => table.table_number === tableNumber || table.qr_token === token);
      if (match) {
        setTableInfo({
          id: match.id,
          tableNumber: match.table_number,
          tableName: match.display_name || `Table ${match.table_number}`,
          qrToken: match.qr_token,
        });
        return;
      }

      setTableInfo({
        id: null,
        tableNumber: tableNumber ?? "",
        tableName: tableNumber ? `Table ${tableNumber}` : "Table",
        qrToken: token ?? "",
      });
    });

    return () => {
      cancelled = true;
    };
  }, [setTableInfo, tableNumber, token]);

  return <MenuPage />;
}

export default function App() {
  return (
    <CartProvider>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/menu" element={<MenuPage />} />
        <Route path="/table/:tableNumber/:token" element={<TableRouteMenuPage />} />
        <Route path="/cart" element={<CartPage />} />
        <Route path="/admin/login" element={<AdminLoginPage />} />
        <Route path="/admin" element={<AdminRoute><AdminDashboardPage /></AdminRoute>} />
        <Route path="/admin/menu" element={<AdminRoute><AdminMenuPage /></AdminRoute>} />
        <Route path="/admin/categories" element={<AdminRoute><AdminCategoriesPage /></AdminRoute>} />
        <Route path="/admin/gallery" element={<AdminRoute><AdminGalleryPage /></AdminRoute>} />
        <Route path="/admin/settings" element={<AdminRoute><AdminSettingsPage /></AdminRoute>} />
        <Route path="/admin/reviews" element={<AdminRoute><AdminReviewsPage /></AdminRoute>} />
        <Route path="/admin/banners" element={<AdminRoute><AdminBannersPage /></AdminRoute>} />
        <Route path="/admin/tables" element={<AdminRoute><AdminTablesPage /></AdminRoute>} />
        <Route path="/admin/kitchen-availability" element={<AdminRoute><AdminKitchenAvailabilityPage /></AdminRoute>} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
      <StickyMobileCartBar />
      <NetworkStatus />
    </CartProvider>
  );
}
