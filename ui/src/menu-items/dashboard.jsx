// assets
import {
  DashboardOutlined,
  DollarOutlined,
  PieChartOutlined,
  WalletOutlined,
  ProfileOutlined,
} from "@ant-design/icons";
// ==============================|| MENU ITEMS - DASHBOARD ||============================== //

const dashboard = {
  id: "group-dashboard",
  title: "My Finances",
  type: "group",
  children: [
    {
      id: "dashboard",
      title: "Dashboard",
      type: "item",
      url: "/dashboard",
      icon: DashboardOutlined,
      breadcrumbs: false,
    },
    {
      id: "budget",
      title: "Budget",
      type: "item",
      url: "/budget",
      icon: PieChartOutlined,
      breadcrumbs: false,
    },
    {
      id: "expenditure",
      title: "Expenditure",
      type: "item",
      url: "/expenditure",
      icon: DollarOutlined,
      breadcrumbs: false,
    },
    {
      id: 'payment-methods',
      title: 'Payment Methods',
      type: 'item',
      url: '/payment-methods',
      icon: WalletOutlined,
      breadcrumbs: false
    },
    {
      id: 'profile',
      title: 'Profile',
      type: 'item',
      url: '/profile',
      icon: ProfileOutlined,
      breadcrumbs: false
    }
  ],
};

export default dashboard;
