// assets
import {CreditCardOutlined} from "@ant-design/icons";

// ==============================|| MENU ITEMS - EXTRA PAGES ||============================== //

const pages = {
  id: "explore",
  title: "Explore",
  type: "group",
  children: [
    {
      id: 'browse-cards',
      title: 'Card Browser',
      type: 'item',
      url: '/browse',
      icon: CreditCardOutlined,
      breadcrumbs: false
    },
  ],
};

export default pages;
