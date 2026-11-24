import { useMemo, useState } from "react";
import ReactApexChart from "react-apexcharts";
import { AGGREGATE_EXPENDITURES, useBudgets } from "../../api/graph";
import { useDateRange } from "../../components/DateRangeProvider";
import { useQuery } from "@apollo/client/react";
import {dateDiff, dateString} from "../../utils/dates";
import {useTheme} from "@mui/material/styles";

const columnChartOptions = {
  chart: {
    type: 'bar',
    height: 450,
    toolbar: {
      show: false
    }
  },
  dataLabels: {
    enabled: false
  },
  yaxis: {
    labels: {
      formatter: (value) => Math.round(value)
    },
    title: {
      text: 'Spending by Category ($)'
    }
  },
};

export default function CategoryChart() {
    const theme = useTheme();
  const [options, setOptions] = useState(columnChartOptions);
  const { startDate, endDate } = useDateRange();
  const { data: expenditures } = useQuery(AGGREGATE_EXPENDITURES, {
    variables: {
      since: dateString(startDate),
      until: dateString(endDate),
      span: "YEAR",
      groupBy: "BUDGET_CATEGORY",
      aggregation: "SUM",
    },
    fetchPolicy: 'network-only'
  });
  const { data: budgets } = useBudgets();

  const [series, setSeries] = useState([]);

    useMemo(() => {
        if (!expenditures?.aggregatedExpenditures || !budgets?.budgets) return;

        // Create map of budget category IDs to expense amounts
        let otherSpendingCategory = "Other";
        const budgetMap = new Map(
            budgets.budgets.flatMap(budget =>
                budget.expenses.map(expense => {
                    if (expense.isSlack) {
                        otherSpendingCategory = expense.category;
                    }
                    return [expense.id, {
                        name: expense.category,
                        amount: expense.amount * (dateDiff(startDate, endDate) + 1) >> 5
                    }]
                })
            )
        );

        const spentMap = new Map(Map.groupBy(expenditures.aggregatedExpenditures, (e) => e.groupByCategory)
            .entries().map(e => [
                budgetMap.has(e[0])? budgetMap.get(e[0]).name : otherSpendingCategory,
                e[1].reduce((acc, curr) => acc + curr.amount, 0)]));

        const budgetedSeries = Array.from(budgetMap.entries())
            .map(([_, value]) => ({ x: value.name, y: value.amount || 0 }))
        // const spentSeries = Array.from(spentMap.entries()).map(([key, value]) => ({ x: key, y: value }))
        const spentSeries = budgetedSeries
            .map((e) => ({x: e.x, y: spentMap.get(e.x) || 0}));

        setSeries([
            {
                name: 'Budgeted',
                data: budgetedSeries
            },
            {
                name: 'Spent',
                data: spentSeries
            }
        ]);

        setOptions((prev) => ({
            ...prev,
            plotOptions: {
                bar: {
                    colors: {
                        ranges: spentSeries.map((spent, index) => ({
                            from: spent.y,
                            to: spent.y,
                            color: spent.y > budgetedSeries[index].y ? theme.palette.error.dark : theme.palette.success.dark
                        }))
                    }
                }
            },
            legend: {
                markers: {
                    fillColors: [null, theme.palette.secondary.main]
                }
            }
        }));
    }, [expenditures?.aggregatedExpenditures, budgets?.budgets, startDate, endDate, theme]);

  return (
    <ReactApexChart
      options={options}
      series={series}
      type="bar"
      height={380}
    />
  );
}