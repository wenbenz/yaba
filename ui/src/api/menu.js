import useSWR, { mutate } from "swr";
import { useMemo } from "react";

const initialState = {
  openedItem: "dashboard",
  openedComponent: "buttons",
  openedHorizontalItem: null,
  isDashboardDrawerOpened: false,
  isComponentDrawerOpened: true,
};

export function useGetMenuMaster() {
  const { data, isLoading } = useSWR("menu", () => initialState, {
    revalidateIfStale: false,
    revalidateOnFocus: false,
    revalidateOnReconnect: false,
  });

  const memoizedValue = useMemo(
    () => ({
      menuMaster: data,
      menuMasterLoading: isLoading,
    }),
    [data, isLoading],
  );

  return memoizedValue;
}

export function handlerDrawerOpen(isDashboardDrawerOpened) {
  // to update local state based on key

  mutate(
    "menu",
    (currentMenuMaster) => {
      return { ...currentMenuMaster, isDashboardDrawerOpened };
    },
    false,
  );
}

export function handlerActiveItem(openedItem) {
  // to update local state based on key

  mutate(
    "menu",
    (currentMenuMaster) => {
      return { ...currentMenuMaster, openedItem };
    },
    false,
  );
}
