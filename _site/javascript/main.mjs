import { generateChart } from "./chart.mjs";
import { dateShortFormatOpt, getTimeFormatOpt, getUrlConfig } from "./util.mjs";

document.addEventListener("alpine:init", () => {
  function root() {
    const today = new Date();
    const oneHourInMs = 60 * 60 * 1000;

    const urlConfig = getUrlConfig();
    const todayRef = new Date(today);
    todayRef.setMinutes(0, 0, 0);

    const daysBefore = new Date(todayRef.getTime() - oneHourInMs);
    const daysAfter = new Date(
      daysBefore.getTime() + urlConfig.end * oneHourInMs,
    );

    return {
      today,
      isMaintenance: false,
      isServerDown: false,
      urlConfig,

      /** @type {Intl.DateTimeFormatOptions} */
      fullDateTimeFormat: {
        ...dateShortFormatOpt,
        ...getTimeFormatOpt(),
      },

      /** @param {{ [key: string]: any }} newParams */
      getUrlWithConfig(newParams) {
        /** @type {typeof newParams} */
        const currentParams = {
          ...getUrlConfig(),
          ...newParams,
        };

        const currentParamsFixed = new URLSearchParams();
        Object.keys(currentParams).forEach((key) => {
          currentParamsFixed.set(key, `${currentParams[key]}`);
        });

        return "?" + currentParamsFixed.toString();
      },

      /** @type {MaintenanceEvent[]} */
      activeEntries: [],

      async init() {
        /** @type {any[]} */
        const res = await fetch("./data.json?v=" + todayRef.getTime()).then(
          (x) => x.json(),
        );

        const allEntries = res.map((entry) => {
          const entryData = Object.keys(entry).reduce((acc, key) => {
            if (key.endsWith("_start") || key.endsWith("_end")) {
              acc[key] = new Date(entry[key]);
            } else {
              acc[key] = entry[key];
            }

            return acc;
          }, /** @type {{ [key: string]: Date | string }} */ ({}));

          return /** @type {MaintenanceEvent} */ (entryData);
        });

        const activeEntries = allEntries.filter(
          (entry) =>
            entry.maintenance_time_start <= today &&
            entry.maintenance_time_end >= today,
        );

        if (activeEntries.length > 0) {
          this.isMaintenance = true;

          this.isServerDown = activeEntries.some(
            (entry) =>
              entry.server_down_start <= today &&
              entry.server_down_end >= today,
          );
        }

        this.activeEntries = activeEntries;
        await generateChart(allEntries, today, daysBefore, daysAfter);
      },
    };
  }

  /** @type {CustomWindow} */ (window).Alpine.data("root", root);
});
