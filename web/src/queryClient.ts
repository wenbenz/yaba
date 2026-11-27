import { Code, ConnectError } from "@connectrpc/connect";
import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query";

const unauthenticated = (error: Error) =>
  ConnectError.from(error).code === Code.Unauthenticated;

export function createQueryClient(router: {
  navigate: (to: string) => unknown;
}) {
  const onError = (error: Error) => {
    if (unauthenticated(error)) router.navigate("/login");
  };
  return new QueryClient({
    queryCache: new QueryCache({ onError }),
    mutationCache: new MutationCache({ onError }),
    defaultOptions: {
      queries: {
        retry: (count, error) => count < 3 && !unauthenticated(error),
      },
    },
  });
}
