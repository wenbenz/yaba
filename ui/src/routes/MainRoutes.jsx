import { lazy } from "react";

// project import
import Loadable from "components/Loadable";
import Dashboard from "layout/Dashboard";

const DashboardDefault = Loadable(lazy(() => import("pages/dashboard/index")));
const BudgetDashboard = Loadable(lazy(() => import("pages/budget/index")));
const ExpenditureDashboard = Loadable(lazy(() => import("pages/expenditure/index")));
const CardBrowser = Loadable(lazy(() => import("pages/browse/index")));
const PaymentMethods = Loadable(lazy(() => import("pages/payment-methods/index")));
const Profile = Loadable(lazy(() => import("pages/profile/index")));
// ==============================|| MAIN ROUTING ||============================== //

const MainRoutes = {
  path: "/",
  element: <Dashboard />,
  children: [
    {
      path: "/",
      element: <DashboardDefault />,
    },
    {
      path: "dashboard",
      element: <DashboardDefault />,
    },
    {
      path: "budget",
      element: <BudgetDashboard />,
    },
    {
      path: "expenditure",
      element: <ExpenditureDashboard />,
    },
    {
      path: "browse",
      element: <CardBrowser />
    },
    {
      path: "payment-methods",
      element: <PaymentMethods />
    },
    {
      path: "profile",
      element: <Profile />
    }
  ],
};

export default MainRoutes;
