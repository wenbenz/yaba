import { useEffect } from "react";
import { Outlet } from "react-router-dom";

// material-ui
import useMediaQuery from "@mui/material/useMediaQuery";
import Toolbar from "@mui/material/Toolbar";
import Box from "@mui/material/Box";

// project import
import Drawer from "./Drawer";
import Header from "./Header";
import navigation from "menu-items";
import Loader from "components/Loader";
import Breadcrumbs from "components/@extended/Breadcrumbs";

import { handlerDrawerOpen, useGetMenuMaster } from "api/menu";

// apollo client
import {
    ApolloClient,
    createHttpLink, from,
    InMemoryCache,
} from "@apollo/client";
import { ApolloProvider } from "@apollo/client/react";
import axios from "axios";
import {onError} from "@apollo/client/link/error";

// ==============================|| MAIN LAYOUT ||============================== //

const httpLink = createHttpLink({
    uri: "/graphql",
    credentials: "include",
});

const errorLink = onError(({ networkError, graphQLErrors }) => {
    if (networkError?.statusCode === 401) {
        console.error("Authentication error:", networkError);
        window.location.href = "/login";
    }
    if (graphQLErrors) {
        graphQLErrors.forEach(({ message, extensions }) => {
            if (extensions?.code === 'UNAUTHENTICATED') {
                console.error("GraphQL authentication error:", message);
                window.location.href = "/login";
            }
        });
    }
});

const client = new ApolloClient({
    link: from([errorLink, httpLink]),
    cache: new InMemoryCache(),
});

// Set up axios interceptors to handle 401 errors
// eslint-disable-next-line no-unused-expressions
axios.interceptors.response.use(
    (response) => {
        if (response.status === 401) {
        window.location.href = "/login";
        }
        return response;
    },
    (error) => {
        if (error.response.status === 401) {
        window.location.href = "/login";
        }
        return Promise.reject(error);
    }
)

export default function DashboardLayout() {
  const { menuMasterLoading } = useGetMenuMaster();
  const downXL = useMediaQuery((theme) => theme.breakpoints.down("xl"));

  useEffect(() => {
    handlerDrawerOpen(!downXL);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [downXL]);

  if (menuMasterLoading) return <Loader />;

  return (
    <Box sx={{ display: "flex", width: "100%" }}>
      <Header />
      <Drawer />
      <Box
        component="main"
        sx={{ width: "calc(100% - 260px)", flexGrow: 1, p: { xs: 2, sm: 3 } }}
      >
        <Toolbar />
        <Breadcrumbs navigation={navigation} title />
        <ApolloProvider client={client}>
          <Outlet />
        </ApolloProvider>
      </Box>
    </Box>
  );
}
