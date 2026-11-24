// material-ui
import Box from "@mui/material/Box";

// third-party
import ReactApexChart from "react-apexcharts";
import Loader from "../../components/Loader";
import {useBudget} from "./BudgetContext";

// chart options
// ==============================|| MONTHLY BAR CHART ||============================== //

export default function BudgetPieChart() {
  // const theme = useTheme();
  const { budget } = useBudget();

  if (!budget) {
    return <Loader />;
  }

  const expenses = budget.expenses.toSorted((a, b) => b.amount - a.amount);
  const labels = expenses.map((e) => e.category);
  const series = expenses.map((e) => e.amount);

  const options = {
    chart: {
      type: "donut",
    },
    plotOptions: {
      pie: {
        donut: {
          labels: {
            show: true,
          },
        },
      },
    },
    labels: labels,
    responsive: [
      {
        breakpoint: 480,
        options: {
          chart: {
            width: 200,
          },
          legend: {
            position: "bottom",
          },
        },
      },
    ],
  };

  return (
    <Box id="chart" sx={{ bgcolor: "transparent" }}>
      <ReactApexChart options={options} series={series} type="donut" />
    </Box>
  );
}
