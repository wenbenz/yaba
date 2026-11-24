export const dateString = (date) => {
  return [
    date.getFullYear(),
    pad(date.getMonth() + 1),
    pad(date.getDate()),
  ].join("-");
};
export const oneWeekAgo = () => {
  let d = new Date();
  d.setDate(d.getDate() - 7);
  return d;
};

export function startOfMonth() {
  let d = new Date();
  d.setFullYear(d.getFullYear(), d.getMonth(), 1);
  return d;
}

export function startOfLastMonth() {
  let d = new Date();
  d.setFullYear(d.getFullYear(), d.getMonth() - 1, 1);
  return d;
}

export function endOfLastMonth() {
  let d = new Date();
  d.setFullYear(d.getFullYear(), d.getMonth(), -1);
  return d;
}

export function startOfYear() {
  let d = new Date();
  d.setFullYear(d.getFullYear(), 0, 1);
  return d;
}

function pad(n) {
  const s = "0" + n;
  return s.substring(s.length - 2);
}

export function getTimeSpan(startDate, endDate) {
  const oneDay = 24 * 60 * 60 * 1000;
  const oneMonth = 30 * oneDay;
  const threeMonths = 3 * oneMonth;
  const threeYears = 3 * 365 * oneDay;

  const timeDifference = endDate - startDate;

  if (timeDifference <= oneMonth) {
    return "DAY";
  } else if (timeDifference <= threeMonths) {
    return "WEEK";
  } else if (timeDifference <= threeYears) {
    return "MONTH";
  } else {
    return "YEAR";
  }
}

export function dateDiff(startDate, endDate) {
  const oneDay = 24 * 60 * 60 * 1000;
  return Math.round((endDate - startDate) / oneDay);
}
