import { useEffect, useState } from "react";
import { fetchDashboard } from "../lib/api";
import type { DashboardSummary } from "../lib/types";

interface ColoniesPageProps {
  empireNo: number;
}

export default function ColoniesPage({ empireNo }: ColoniesPageProps) {
  const [data, setData] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchDashboard(empireNo)
      .then(setData)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [empireNo]);

  if (loading) return <p className="text-gray-500">Loading…</p>;
  if (error) return <p className="text-red-600">{error}</p>;
  if (!data || data.colony_count === 0) {
    return (
      <div>
        <h1 className="text-2xl font-semibold text-gray-900 mb-4">Colonies</h1>
        <p className="text-gray-500">No colonies.</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900 mb-4">Colonies</h1>
      <p className="text-sm text-gray-500 mb-4">
        {data.colony_count} {data.colony_count === 1 ? "colony" : "colonies"} total
      </p>
      <table className="min-w-full divide-y divide-gray-200">
        <thead>
          <tr>
            <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">Kind</th>
            <th className="px-4 py-2 text-right text-sm font-medium text-gray-500">Count</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {data.colony_kinds.map((kc) => (
            <tr key={kc.kind}>
              <td className="px-4 py-2 text-sm text-gray-900">{kc.kind}</td>
              <td className="px-4 py-2 text-sm text-gray-900 text-right">{kc.count}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
