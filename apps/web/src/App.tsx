import { useEffect, useState } from "react";
import {
  HomeIcon,
  DocumentTextIcon,
  ClipboardDocumentListIcon,
} from "@heroicons/react/24/outline";
import { clearToken } from "./lib/auth";
import { fetchMe, loginWithMagicLink } from "./lib/api";
import AppShell from "./components/AppShell";
import AdminUsersPage from "./pages/AdminUsersPage";
import DashboardPage from "./pages/DashboardPage";
import OrdersPage from "./pages/OrdersPage";
import ReportsPage from "./pages/ReportsPage";
import ReportPage from "./pages/ReportPage";
import type { NavItem } from "./components/AppShell";

type Page = "dashboard" | "orders" | "reports" | "report" | "admin-users";

function ClusterUnderConstruction() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950">
      <div className="text-center px-6">
        <p className="text-indigo-400 text-sm font-mono uppercase tracking-widest mb-4">
          Epimethean Challenge
        </p>
        <h1 className="text-4xl font-bold text-white mb-3">
          Cluster Under Construction
        </h1>
        <p className="text-gray-400 max-w-sm mx-auto">
          The cluster is not yet open to travelers. Check your mission briefing
          for access instructions.
        </p>
      </div>
    </div>
  );
}

function App() {
  const [authenticated, setAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const [empireNo, setEmpireNo] = useState(0);
  const [empireName, setEmpireName] = useState("");
  const [page, setPage] = useState<Page>("dashboard");
  const [reportLink, setReportLink] = useState("");

  useEffect(() => {
    async function init() {
      // Handle magic link in URL: ?magic=<token>
      const params = new URLSearchParams(window.location.search);
      const magic = params.get("magic");
      if (magic) {
        try {
          await loginWithMagicLink(magic);
          // Remove magic param from URL without reload
          params.delete("magic");
          const newSearch = params.toString();
          const newUrl = window.location.pathname + (newSearch ? `?${newSearch}` : "");
          window.history.replaceState({}, "", newUrl);
        } catch {
          // Invalid magic link — fall through to /me check
        }
      }

      // Check auth status via /api/me
      try {
        const me = await fetchMe();
        if (me.authenticated) {
          setEmpireNo(me.empire);
          setEmpireName(me.name);
          setAuthenticated(true);
        }
      } catch {
        // Network error or server down — stay unauthenticated
      }
      setLoading(false);
    }
    init();
  }, []);

  function handleSignOut() {
    clearToken();
    setAuthenticated(false);
    setEmpireNo(0);
    setEmpireName("");
    setPage("dashboard");
  }

  const userNavigation = [
    { name: "Sign out", href: "#", onClick: handleSignOut },
  ];

  const navigation: NavItem[] = [
    {
      name: "Dashboard",
      href: "#",
      icon: HomeIcon,
      current: page === "dashboard",
      onClick: () => setPage("dashboard"),
    },
    {
      name: "Orders",
      href: "#",
      icon: ClipboardDocumentListIcon,
      current: page === "orders",
      onClick: () => setPage("orders"),
    },
    {
      name: "Reports",
      href: "#",
      icon: DocumentTextIcon,
      current: page === "reports" || page === "report",
      onClick: () => setPage("reports"),
    },
  ];

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-950">
        <p className="text-gray-400 font-mono text-sm">Connecting…</p>
      </div>
    );
  }

  if (!authenticated) {
    return <ClusterUnderConstruction />;
  }

  function renderPage() {
    switch (page) {
      case "orders":
        return <OrdersPage empireNo={empireNo} />;
      case "reports":
        return (
          <ReportsPage
            empireNo={empireNo}
            onSelectReport={(link) => {
              setReportLink(link);
              setPage("report");
            }}
          />
        );
      case "report":
        return (
          <ReportPage link={reportLink} onBack={() => setPage("reports")} />
        );
      case "admin-users":
        return <AdminUsersPage />;
      default:
        return (
          <DashboardPage
            empireName={empireName}
            onNavigateOrders={() => setPage("orders")}
            onNavigateReports={() => setPage("reports")}
          />
        );
    }
  }

  return (
    <AppShell
      navigation={navigation}
      userNavigation={userNavigation}
      userName={empireName}
    >
      {renderPage()}
    </AppShell>
  );
}

export default App;
