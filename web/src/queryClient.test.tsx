import { Code, ConnectError, createRouterTransport } from "@connectrpc/connect";
import { TransportProvider, useQuery } from "@connectrpc/connect-query";
import { QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { PaymentMethodService } from "./gen/yaba/v1/payment_method_pb";
import { createQueryClient } from "./queryClient";

it("navigates to /login when a query fails with Unauthenticated", async () => {
  const navigate = vi.fn();
  const queryClient = createQueryClient({ navigate });
  const transport = createRouterTransport((router) =>
    router.service(PaymentMethodService, {
      listPaymentMethods: () => {
        throw new ConnectError("no session", Code.Unauthenticated);
      },
    }),
  );

  renderHook(() => useQuery(PaymentMethodService.method.listPaymentMethods), {
    wrapper: ({ children }) => (
      <TransportProvider transport={transport}>
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      </TransportProvider>
    ),
  });

  await waitFor(() => expect(navigate).toHaveBeenCalledWith("/login"));
});
