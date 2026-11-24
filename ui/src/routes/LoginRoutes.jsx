import { lazy } from "react";

// project import
import Loadable from "components/Loadable";
import MinimalLayout from "layout/MinimalLayout";

// render - login
const AuthLogin = Loadable(lazy(() => import("pages/authentication/login")));
const AuthRegister = Loadable(
  lazy(() => import("pages/authentication/register")),
);
const TermsOfService = Loadable(lazy(() => import("pages/terms")));
const PrivacyPolicy = Loadable(lazy(() => import("pages/privacy")));

// ==============================|| AUTH ROUTING ||============================== //

const LoginRoutes = {
  path: "/",
  element: <MinimalLayout />,
  children: [
    {
      path: "/login",
      element: <AuthLogin />,
    },
    {
      path: "/register",
      element: <AuthRegister />,
    },
    {
      path: "/terms",
      element: <TermsOfService />,
    },
    {
      path: "/privacy",
      element: <PrivacyPolicy />,
    },
  ],
};

export default LoginRoutes;
