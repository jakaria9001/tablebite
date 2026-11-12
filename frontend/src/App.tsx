import { useEffect } from "react";
import { Route, Routes, useLocation, useParams } from "react-router-dom";
import { StickyMobileCartBar } from "./components/cart/StickyMobileCartBar";
import { CartProvider } from "./context/CartContext";
import { useCart } from "./hooks/useCart";
import { resolveTableSession } from "./api/tableSessions";
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
import OrderReviewPage from "./pages/OrderReviewPage";
import OrderDetailsPage from "./pages/OrderDetailsPage";
import WaiterLoginPage from "./pages/WaiterLoginPage";
import WaiterDashboardPage from "./pages/WaiterDashboardPage";
import { AdminRoute } from "./components/common/AdminRoute";
import { NetworkStatus } from "./components/common/NetworkStatus";

function TableRouteMenuPage() {
  const { token } = useParams();
  const { setTableInfo } = useCart();

  useEffect(() => {
    if (!token) {
      setTableInfo(null);
      return;
    }

    let cancelled = false;

    void resolveTableSession(token).then((session) => {
      if (cancelled) {
        return;
      }

      setTableInfo({
        id: session.tableId,
        sessionId: session.id,
        tableNumber: session.tableNumber,
        tableName: session.tableName || `Table ${session.tableNumber}`,
        qrToken: session.qrToken,
      });
    }).catch(() => {
      if (!cancelled) {
        setTableInfo(null);
      }
    });

    return () => {
      cancelled = true;
    };
  }, [setTableInfo, token]);

  return <MenuPage />;
}

function ScrollToTop() {
  const { pathname } = useLocation();

  useEffect(() => {
    window.scrollTo({ top: 0, left: 0, behavior: "auto" });
  }, [pathname]);

  return null;
}

export default function App() {
  return (
    <CartProvider>
      <ScrollToTop />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/menu" element={<MenuPage />} />
        <Route path="/table/:tableNumber/:token" element={<TableRouteMenuPage />} />
        <Route path="/cart" element={<CartPage />} />
        <Route path="/order/review" element={<OrderReviewPage />} />
        <Route path="/orders/:orderId" element={<OrderDetailsPage />} />
        <Route path="/waiter/login" element={<WaiterLoginPage />} />
        <Route path="/waiter" element={<WaiterDashboardPage />} />
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
