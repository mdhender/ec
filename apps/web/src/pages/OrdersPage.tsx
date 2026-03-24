import { useEffect, useState } from "react";
import { fetchOrders, submitOrders } from "../lib/api";

interface OrdersPageProps {
  empireNo: number;
}

export default function OrdersPage({ empireNo }: OrdersPageProps) {
  const [orders, setOrders] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    fetchOrders(empireNo)
      .then(setOrders)
      .catch((err: Error) => {
        // 404 means no orders yet — start with empty
        if (err.message.includes("404") || err.message.includes("not found")) {
          setOrders("");
        } else {
          setError(err.message);
        }
      })
      .finally(() => setLoading(false));
  }, [empireNo]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setSaved(false);
    try {
      await submitOrders(empireNo, orders);
      setSaved(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save orders");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return <p className="text-gray-500">Loading orders…</p>;
  }

  return (
    <div className="max-w-2xl">
      <h1 className="text-2xl font-semibold text-gray-900 mb-4">Orders</h1>
      {error && (
        <p className="mb-4 text-sm text-red-600">{error}</p>
      )}
      <form onSubmit={handleSubmit}>
        <textarea
          value={orders}
          onChange={(e) => { setOrders(e.target.value); setSaved(false); }}
          rows={20}
          className="w-full font-mono text-sm border border-gray-300 rounded-md p-3 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="Enter your orders here…"
        />
        <div className="mt-3 flex items-center gap-4">
          <button
            type="submit"
            disabled={saving}
            className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50 transition-colors"
          >
            {saving ? "Saving…" : "Save Orders"}
          </button>
          {saved && <span className="text-sm text-green-600">Saved.</span>}
        </div>
      </form>
    </div>
  );
}
