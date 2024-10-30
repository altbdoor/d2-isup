/** @return {Intl.DateTimeFormatOptions} */
export const getTimeFormatOpt = () => {
  return {
    hour: "numeric",
    minute: "numeric",
    hour12: getUrlConfig().mode === "12h",
  };
};

/** @type {Intl.DateTimeFormatOptions} */
export const dateShortFormatOpt = {
  day: "numeric",
  month: "short",
  year: "numeric",
};

export const getUrlConfig = () => {
  const url = new URL(window.location.href);

  return {
    mode: url.searchParams.get("mode") || "12h",
    end: parseInt(url.searchParams.get("end") || "48", 10),
  };
};
