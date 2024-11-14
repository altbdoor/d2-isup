import { getAdapter } from "./chart-date-adapter.mjs";
import { dateShortFormatOpt, getTimeFormatOpt } from "./util.mjs";

/**
  @param {MaintenanceEvent[]} allEntries
  @param {Date} today
  @param {Date} daysBefore
  @param {Date} daysAfter
 */
export const generateChart = async (
  allEntries,
  today,
  daysBefore,
  daysAfter,
) => {
  const timeFormatOpt = getTimeFormatOpt();
  const darkMode = window.matchMedia("(prefers-color-scheme: dark)").matches;

  const datasets = [
    {
      label: "Maintenance window",
      order: 30,
      borderWidth: 2,
      borderColor: "rgb(255, 205, 86)",
      backgroundColor: "rgba(255, 205, 86, 0.6)",
      data: allEntries.map((entry) => ({
        y: 0,
        x: [
          entry.maintenance_time_start.getTime(),
          entry.maintenance_time_end.getTime(),
        ],
      })),
    },
    {
      label: "Server down",
      order: 20,
      borderWidth: 2,
      borderColor: "rgb(255, 99, 132)",
      backgroundColor: "rgba(255, 99, 132, 0.6)",
      data: allEntries.map((entry) => ({
        y: 0,
        x: [entry.server_down_start.getTime(), entry.server_down_end.getTime()],
      })),
    },
    {
      label: "Now",
      order: 10,
      inflateAmount: 2,
      backgroundColor: "rgb(54, 162, 235)",
      data: [
        {
          y: 0,
          x: [today.getTime(), today.getTime()],
        },
      ],
    },
  ];

  const tooltipPlugin = {
    xAlign: "right",
    yAlign: "bottom",
    caretSize: 0,
    bodyFont: { size: 14 },
    callbacks: {
      title: () => "",
      label: (item) => {
        const xVals = item.raw.x;
        if (xVals.length !== 2) {
          return "";
        }

        const startDate = new Date(xVals[0]);
        const endDate = new Date(xVals[1]);

        return [
          startDate.toLocaleString(undefined, timeFormatOpt),
          "-",
          endDate.toLocaleString(undefined, timeFormatOpt),
        ].join(" ");
      },
    },
  };

  const chartComponents = [
    "BarController",
    "BarElement",
    "Legend",
    "Tooltip",
    "TimeScale",
    "CategoryScale",
  ];

  const chartImportUrl = new URL("https://esm.sh/chart.js@4.4.6");
  chartImportUrl.searchParams.set("bundle-deps", "");
  chartImportUrl.searchParams.set(
    "exports",
    [...chartComponents, "_adapters", "Chart"].join(","),
  );

  // @ts-ignore
  const chartlib = await import(chartImportUrl.toString());

  chartComponents.forEach((chartComp) => {
    chartlib.Chart.register(chartlib[chartComp]);
  });

  chartlib._adapters._date.override(await getAdapter());

  chartlib.Chart.defaults.color = darkMode ? "#fff" : "#666";
  chartlib.Chart.defaults.borderColor = darkMode
    ? "rgba(255, 255, 255, 0.3)"
    : "rgba(0, 0, 0, 0.1)";

  const chart = new chartlib.Chart(document.getElementById("chart"), {
    type: "bar",
    data: { datasets },
    options: {
      animation: false,
      indexAxis: "y",
      maintainAspectRatio: false,
      resizeDelay: 100,
      scales: {
        x: {
          type: "time",
          position: "top",
          min: daysBefore.getTime(),
          max: daysAfter.getTime(),
          ticks: {
            autoSkip: true,
            align: "start",
            maxRotation: 40,
            major: { enabled: true },

            /** @param {number} value */
            callback: (value) => {
              const tmpDate = new Date(value);
              if (tmpDate.getHours() === 0) {
                return tmpDate.toLocaleString(undefined, dateShortFormatOpt);
              }

              return tmpDate.toLocaleString(undefined, timeFormatOpt);
            },
            font: (ctx) => {
              const tmpDate = new Date(ctx.tick.value);
              if (tmpDate.getHours() === 0) {
                return { weight: "bold", size: 14 };
              }

              return {};
            },
          },
          time: { unit: "hour" },
        },
        y: { stacked: true },
      },
      elements: {
        bar: {
          borderSkipped: false,
        },
      },
      plugins: {
        legend: { position: "bottom" },
        tooltip: tooltipPlugin,
      },
    },
  });

  return chart;
};
