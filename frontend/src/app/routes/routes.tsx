import { lazy } from "react";

import { ROUTES } from "./constants";
import { createBrowserRouter } from "react-router";

const Dashboard = lazy(() => import("@pages/dashboard/dashboard"));
const ServiceDetail = lazy(() => import("@/pages/service/serviceDetail"));
const NotFound = lazy(() => import("@/pages/notFound/notFound"));

const routes = [
  { path: ROUTES.DASHBOARD, element: <Dashboard /> },
  { path: ROUTES.SERVICE_DETAIL, element: <ServiceDetail /> },
  { path: ROUTES.NOT_FOUND, element: <NotFound /> },
];

export const router = createBrowserRouter(routes);
