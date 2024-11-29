/**
  reference: https://github.com/bolstycjw/chartjs-adapter-dayjs-4 (MIT)
  reference: https://github.com/chartjs/chartjs-adapter-luxon (MIT)

  edited to dynamically load dayjs without any formats, because all the date
  formatting is handled within `chart.mjs` directly.
 */

const FORMATS = {};

export const getAdapter = async () => {
  // @ts-ignore
  const mod = await import("https://cdn.jsdelivr.net/npm/dayjs@1.11.13/+esm");
  const dayjs = mod.default;

  return {
    _id: "custom-dayjs", // DEBUG
    formats: function () {
      return FORMATS;
    },
    parse: function parse(value, format) {
      var valueType = typeof value;
      if (value === null || valueType === "undefined") {
        return null;
      }
      if (valueType === "string" && typeof format === "string") {
        return dayjs(value, format).isValid()
          ? dayjs(value, format).valueOf()
          : null;
      } else if (!(value instanceof dayjs)) {
        return dayjs(value).isValid() ? dayjs(value).valueOf() : null;
      }
      return null;
    },
    format: function format(time, _format) {
      return dayjs(time).format(_format);
    },
    add: function add(time, amount, unit) {
      return dayjs(time).add(amount, unit).valueOf();
    },
    diff: function diff(max, min, unit) {
      return dayjs(max).diff(dayjs(min), unit);
    },
    startOf: function startOf(time, unit, weekday) {
      if (unit === "isoWeek") {
        // Ensure that weekday has a valid format
        var validatedWeekday =
          typeof weekday === "number" && weekday > 0 && weekday < 7
            ? weekday
            : 1;
        return dayjs(time)
          .isoWeekday(validatedWeekday)
          .startOf("day")
          .valueOf();
      }
      return dayjs(time).startOf(unit).valueOf();
    },
    endOf: function endOf(time, unit) {
      return dayjs(time).endOf(unit).valueOf();
    },
  };
};
