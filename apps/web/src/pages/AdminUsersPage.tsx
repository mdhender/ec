import { useEffect, useState, type FormEvent } from "react";
import { fetchUsers, createUser } from "../lib/api";
import type { UserSummary } from "../lib/types";

export default function AdminUsersPage() {
  const [users, setUsers] = useState<UserSummary[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("player");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  function loadUsers() {
    fetchUsers()
      .then(setUsers)
      .catch((err) => {
        setError(err instanceof Error ? err.message : "Failed to load users");
      });
  }

  useEffect(() => {
    loadUsers();
  }, []);

  function resetForm() {
    setError("");
    setUsername("");
    setEmail("");
    setPassword("");
    setRole("player");
  }

  function handleToggleForm() {
    setShowForm((v) => !v);
    resetForm();
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await createUser({ username, email, password, role });
      loadUsers();
      setShowForm(false);
      resetForm();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create user");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold text-gray-900">Users</h1>
        <button
          type="button"
          onClick={handleToggleForm}
          className="rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
          {showForm ? "Cancel" : "Create User"}
        </button>
      </div>

      {error && (
        <p className="mb-4 rounded-md bg-red-50 px-4 py-3 text-sm text-red-600">
          {error}
        </p>
      )}

      {showForm && (
        <form
          onSubmit={handleSubmit}
          className="mb-6 rounded-md border border-gray-200 bg-white p-4 shadow-xs space-y-4"
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label
                htmlFor="new-username"
                className="block text-sm font-medium text-gray-900"
              >
                Username
              </label>
              <input
                id="new-username"
                type="text"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="mt-1 block w-full rounded-md bg-white px-3 py-1.5 text-sm text-gray-900 outline-1 -outline-offset-1 outline-gray-300 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600"
              />
            </div>
            <div>
              <label
                htmlFor="new-email"
                className="block text-sm font-medium text-gray-900"
              >
                Email
              </label>
              <input
                id="new-email"
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 block w-full rounded-md bg-white px-3 py-1.5 text-sm text-gray-900 outline-1 -outline-offset-1 outline-gray-300 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600"
              />
            </div>
            <div>
              <label
                htmlFor="new-password"
                className="block text-sm font-medium text-gray-900"
              >
                Password
              </label>
              <input
                id="new-password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 block w-full rounded-md bg-white px-3 py-1.5 text-sm text-gray-900 outline-1 -outline-offset-1 outline-gray-300 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600"
              />
            </div>
            <div>
              <label
                htmlFor="new-role"
                className="block text-sm font-medium text-gray-900"
              >
                Role
              </label>
              <select
                id="new-role"
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="mt-1 block w-full rounded-md bg-white px-3 py-1.5 text-sm text-gray-900 outline-1 -outline-offset-1 outline-gray-300 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600"
              >
                <option value="player">Player</option>
                <option value="admin">Admin</option>
              </select>
            </div>
          </div>

          <div className="flex justify-end">
            <button
              type="submit"
              disabled={loading}
              className="rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50"
            >
              {loading ? "Creating\u2026" : "Create"}
            </button>
          </div>
        </form>
      )}

      <div className="overflow-hidden rounded-md border border-gray-200 bg-white shadow-xs">
        <table className="min-w-full divide-y divide-gray-200">
          <thead>
            <tr>
              <th
                scope="col"
                className="px-4 py-3 text-left text-sm font-semibold text-gray-900"
              >
                Username
              </th>
              <th
                scope="col"
                className="px-4 py-3 text-left text-sm font-semibold text-gray-900"
              >
                Email
              </th>
              <th
                scope="col"
                className="px-4 py-3 text-left text-sm font-semibold text-gray-900"
              >
                Role
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {users.map((user) => (
              <tr key={user.user_id}>
                <td className="px-4 py-3 text-sm text-gray-900">
                  {user.username}
                </td>
                <td className="px-4 py-3 text-sm text-gray-900">
                  {user.email}
                </td>
                <td className="px-4 py-3 text-sm text-gray-900">
                  {user.role}
                </td>
              </tr>
            ))}
            {users.length === 0 && (
              <tr>
                <td
                  colSpan={3}
                  className="px-4 py-6 text-center text-sm text-gray-500"
                >
                  No users found.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
