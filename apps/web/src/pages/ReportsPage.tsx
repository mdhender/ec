import { useEffect, useState } from "react";
import { fetchReports } from "../lib/api";
import type { ReportSummary } from "../lib/types";

interface ReportsPageProps {
  empireNo: number;
  onSelectReport: (link: string) => void;
}

export default function ReportsPage({ empireNo, onSelectReport }: ReportsPageProps) {
  const [reports, setReports] = useState<ReportSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchReports(empireNo)
      .then(setReports)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [empireNo]);

  if (loading) return <p className="text-gray-500">Loading reports…</p>;
  if (error) return <p className="text-sm text-red-600">{error}</p>;

  return (
    <div className="max-w-xl">
      <h1 className="text-2xl font-semibold text-gray-900 mb-4">Reports</h1>
      {reports.length === 0 ? (
        <p className="text-gray-500">No reports are available yet.</p>
      ) : (
        <ul className="space-y-2">
          {reports.map((r) => (
            <li key={r.link}>
              <button
                onClick={() => onSelectReport(r.link)}
                className="text-indigo-600 hover:underline"
              >
                Year {r.turn_year}, Quarter {r.turn_quarter}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
