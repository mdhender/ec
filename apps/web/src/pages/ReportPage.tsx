import { useEffect, useState } from "react";
import { fetchReportByLink } from "../lib/api";

interface ReportPageProps {
  link: string;
  onBack: () => void;
}

export default function ReportPage({ link, onBack }: ReportPageProps) {
  const [content, setContent] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchReportByLink(link)
      .then((data) => setContent(JSON.stringify(data, null, 2)))
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [link]);

  if (loading) return <p className="text-gray-500">Loading report…</p>;
  if (error) return <p className="text-sm text-red-600">{error}</p>;

  return (
    <div className="max-w-4xl">
      <div className="mb-4">
        <button
          onClick={onBack}
          className="text-sm text-indigo-600 hover:underline"
        >
          ← Back to reports
        </button>
      </div>
      <h1 className="text-2xl font-semibold text-gray-900 mb-4">Report</h1>
      <pre className="font-mono text-sm bg-gray-50 border border-gray-200 rounded-md p-4 overflow-auto whitespace-pre-wrap">
        {content}
      </pre>
    </div>
  );
}
