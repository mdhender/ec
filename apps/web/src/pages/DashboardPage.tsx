import { useEffect, useState } from "react";
import { fetchDashboard } from "../lib/api";
import type { DashboardSummary, KindCount } from "../lib/types";

interface DashboardPageProps {
  empireName: string;
  empireNo: number;
  onNavigateOrders: () => void;
  onNavigateReports: () => void;
  onNavigateColonies: () => void;
  onNavigateShips: () => void;
}

interface SummaryCardProps {
  title: string;
  count: number;
  kinds: KindCount[];
  onNavigate?: () => void;
}

function SummaryCard({ title, count, kinds, onNavigate }: SummaryCardProps) {
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h3 className="text-sm font-medium text-gray-500 uppercase tracking-wide">
        {title}
      </h3>
      <p className="text-3xl font-bold text-gray-900 mt-2">{count}</p>
      <ul className="mt-2 space-y-1">
        {kinds.map((kc) => (
          <li key={kc.kind} className="text-sm text-gray-600">
            {kc.count} {kc.kind}
          </li>
        ))}
      </ul>
      {onNavigate && (
        <button
          onClick={onNavigate}
          className="mt-4 text-sm text-indigo-600 hover:text-indigo-800 font-medium"
        >
          View details →
        </button>
      )}
    </div>
  );
}

export default function DashboardPage({
  empireName,
  empireNo,
  onNavigateOrders,
  onNavigateReports,
  onNavigateColonies,
  onNavigateShips,
}: DashboardPageProps) {
  const [data, setData] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchDashboard(empireNo)
      .then(setData)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [empireNo]);

  return (
    <div className="max-w-3xl">
      <h1 className="text-2xl font-semibold text-gray-900 mb-6">{empireName}</h1>

      {error && <p className="text-sm text-red-600 mb-4">{error}</p>}

      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
          {[0, 1, 2].map((i) => (
            <div key={i} className="bg-white rounded-lg shadow p-6 animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-1/2 mb-3" />
              <div className="h-8 bg-gray-200 rounded w-1/4" />
            </div>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
          <SummaryCard
            title="Colonies"
            count={data?.colony_count ?? 0}
            kinds={data?.colony_kinds ?? []}
            onNavigate={onNavigateColonies}
          />
          <SummaryCard
            title="Ships"
            count={data?.ship_count ?? 0}
            kinds={[]}
            onNavigate={onNavigateShips}
          />
          <SummaryCard
            title="Planets"
            count={data?.planet_count ?? 0}
            kinds={data?.planet_kinds ?? []}
          />
        </div>
      )}

      <div className="flex gap-4">
        <button
          onClick={onNavigateOrders}
          className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition-colors"
        >
          Orders
        </button>
        <button
          onClick={onNavigateReports}
          className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition-colors"
        >
          Reports
        </button>
      </div>
    </div>
  );
}
