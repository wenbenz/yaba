import { NavLink, Outlet } from "react-router";

const sections = [
  ["Dashboard", "/"],
  ["Budget", "/budget"],
  ["Expenditure", "/expenditure"],
  ["Payment methods", "/payment-methods"],
  ["Reward cards", "/reward-cards"],
  ["Profile", "/profile"],
] as const;

export default function Shell() {
  return (
    <div className="flex min-h-screen">
      <aside className="w-56 border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
        <nav className="flex flex-col gap-1 p-4">
          {sections.map(([label, to]) => (
            <NavLink
              key={to}
              to={to}
              end={to === "/"}
              className="rounded-md px-3 py-2 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground aria-[current=page]:bg-sidebar-accent aria-[current=page]:text-sidebar-accent-foreground"
            >
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <div className="flex flex-1 flex-col">
        <header className="border-b border-border px-6 py-4 font-semibold">
          YABA
        </header>
        <main className="flex-1 p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
