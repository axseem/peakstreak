import type { HabitLog } from "../types";

/**
 * Formats a Date object into a YYYY-MM-DD string.
 */
export const toYYYYMMDD = (date: Date): string => date.toISOString().split('T')[0];

/**
 * Generates an array of Date objects for every day in a given year (UTC).
 * For the current year, it only generates dates up to the current day.
 * For future years, it returns an empty array.
 */
export const getDatesForYear = (year: number): Date[] => {
  const dates: Date[] = [];
  const date = new Date(Date.UTC(year, 0, 1));

  const now = new Date();
  const currentYear = now.getUTCFullYear();

  if (year > currentYear) {
    return [];
  }

  const isCurrentYear = year === currentYear;
  const today = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()));

  while (date.getUTCFullYear() === year) {
    if (isCurrentYear && date > today) {
      break;
    }
    dates.push(new Date(date));
    date.setUTCDate(date.getUTCDate() + 1);
  }
  return dates;
};

/**
 * Groups habit logs by their year.
 */
export const groupLogsByYear = (logs: HabitLog[]): Record<number, HabitLog[]> => {
  return logs.reduce((acc, log) => {
    const year = new Date(log.date).getUTCFullYear();
    if (!acc[year]) {
      acc[year] = [];
    }
    acc[year].push(log);
    return acc;
  }, {} as Record<number, HabitLog[]>);
};
