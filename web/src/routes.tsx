import type { RouteObject } from "react-router";
import Shell from "./Shell";

export const routes: RouteObject[] = [
  { path: "/login", element: <h1>Sign in</h1> },
  {
    path: "/",
    element: <Shell />,
    children: [{ index: true, element: <h1>Dashboard</h1> }],
  },
];
