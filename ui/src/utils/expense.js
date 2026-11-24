import { dateDiff } from "./dates";

export function expenseIdToName(budget, id) {
  const uuid2category = new Map(
    budget.expenses.map((expense) => [expense.id, expense.category]),
  );
  const misc = budget.expenses.filter((expense) => expense.isSlack)[0];
  return uuid2category.get(id) || misc.category;
}

export function expenseNameToId(budget, name) {
  const category2uuid = new Map(
    budget.expenses.map((expense) => [expense.category, expense.id]),
  );
  const misc = budget.expenses.filter((expense) => expense.isSlack)[0];
  return category2uuid.get(name) || misc.id;
}

export function calculateSpan(startDate, endDate) {
  const daysDiff = dateDiff(startDate, endDate);

  if (daysDiff <= 31) return "DAY";
  if (daysDiff <= 210) return "WEEK";
  return "MONTH";
}

export function adjustBudgetForSpan(monthlyBudget, span) {
  switch (span) {
    case "DAY":
      return monthlyBudget / 30; // Approximate daily budget
    case "WEEK":
      return (monthlyBudget * 12) / 52; // Weekly budget
    default:
      return monthlyBudget;
  }
}
