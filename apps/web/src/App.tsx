import { useEffect, useState } from "react";
import { HomeIcon, UsersIcon } from "@heroicons/react/24/outline";
import { getToken, clearToken } from "./lib/auth";
import { fetchProfile } from "./lib/api";
import AppShell from "./components/AppShell";
import LoginPage from "./pages/LoginPage";
import AdminUsersPage from "./pages/AdminUsersPage";
import type { NavItem } from "./components/AppShell";
import type { Profile } from "./lib/types";

type Page = "dashboard" | "admin-users";

function App() {
  const [authenticated, setAuthenticated] = useState(() => !!getToken());
  const [profile, setProfile] = useState<Profile | null>(null);
  const [page, setPage] = useState<Page>("dashboard");

  useEffect(() => {
    if (!authenticated) return;
    fetchProfile()
      .then(setProfile)
      .catch(() => {
        clearToken();
        setAuthenticated(false);
      });
  }, [authenticated]);

  function handleSignOut() {
    clearToken();
    setProfile(null);
    setPage("dashboard");
    setAuthenticated(false);
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
    ...(profile?.role === "admin"
      ? [
          {
            name: "Users",
            href: "#",
            icon: UsersIcon,
            current: page === "admin-users",
            onClick: () => setPage("admin-users"),
          } satisfies NavItem,
        ]
      : []),
  ];

  if (!authenticated) {
    return <LoginPage onLogin={() => setAuthenticated(true)} />;
  }

  return (
    <AppShell
      navigation={navigation}
      userNavigation={userNavigation}
      userName={profile?.handle ?? ""}
    >
      {page === "admin-users" ? (
        <AdminUsersPage />
      ) : (
        <p className="text-lg text-gray-700">Dashboard</p>
      )}
    </AppShell>
  );
}

export default App;
