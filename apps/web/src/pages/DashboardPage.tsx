interface DashboardPageProps {
  empireName: string;
  onNavigateOrders: () => void;
  onNavigateReports: () => void;
}

export default function DashboardPage({
  empireName,
  onNavigateOrders,
  onNavigateReports,
}: DashboardPageProps) {
  return (
    <div className="max-w-xl">
      <h1 className="text-2xl font-semibold text-gray-900 mb-6">{empireName}</h1>
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
