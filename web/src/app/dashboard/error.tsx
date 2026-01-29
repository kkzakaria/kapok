"use client";

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex min-h-[400px] items-center justify-center">
      <div className="text-center">
        <h2 className="text-lg font-semibold text-gray-900">Failed to load dashboard</h2>
        <p className="mt-2 text-sm text-gray-500">{error.message}</p>
        <button
          onClick={reset}
          className="mt-4 rounded-lg bg-kapok-600 px-4 py-2 text-sm font-medium text-white hover:bg-kapok-700"
        >
          Try again
        </button>
      </div>
    </div>
  );
}
