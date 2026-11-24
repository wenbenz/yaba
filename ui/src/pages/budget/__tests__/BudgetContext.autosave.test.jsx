/**
 * Unit tests for BudgetContext auto-save logic.
 * Covers: TC-001–TC-007, TC-018–TC-024
 *
 * Skip-flag mechanics recap:
 *   skipNextSaveRef starts true. The template budget has no `id`, so Effect 2
 *   returns early without touching the ref. When a budget WITH an id is loaded
 *   via setBudget (simulating server data arrival), Effect 2 fires, sees the
 *   flag is true, flips it to false, and returns — that change is skipped.
 *   Only the NEXT id-bearing change (a real user edit) proceeds to save.
 */

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import React from 'react';

// ---------------------------------------------------------------------------
// Mock ../../api/graph BEFORE importing anything that depends on it.
// vi.mock calls are hoisted to the top of the module by Vitest.
// ---------------------------------------------------------------------------
const updateBudgetMock = vi.fn();

vi.mock('../../../api/graph', () => ({
  useBudgets: vi.fn(),
  useUpdateBudget: vi.fn(),
}));

vi.mock('../../../components/Loader', () => ({ default: () => <div data-testid="loader" /> }));
vi.mock('../CreateBudget', () => ({ default: () => <div data-testid="create-budget" /> }));

// ---------------------------------------------------------------------------
// Import the modules under test after mocks are registered.
// ---------------------------------------------------------------------------
import { useBudgets, useUpdateBudget } from '../../../api/graph';
import { BudgetProvider, useBudget } from '../BudgetContext';

// ---------------------------------------------------------------------------
// Shared fixtures
// ---------------------------------------------------------------------------

const LOADED_BUDGET = {
  id: '1',
  name: 'Test Budget',
  incomes: [{ source: 'Work', amount: 2400 }],
  expenses: [{ category: 'Rent', amount: 1000, isFixed: true, isSlack: false }],
};

/**
 * Minimal harness that exposes setBudget and saveStatus via the DOM.
 * The button calls setBudget with a hard-coded id-bearing edit object.
 */
function Harness() {
  const { setBudget, saveStatus } = useBudget();
  return (
    <div>
      <span data-testid="status">{saveStatus}</span>
      <button
        data-testid="edit-btn"
        onClick={() =>
          setBudget({ id: '1', name: 'Edited', incomes: [], expenses: [] })
        }
      >
        Edit
      </button>
    </div>
  );
}

/**
 * Configures mocks and renders BudgetProvider with Harness inside.
 * useBudgets returns server data immediately, causing the provider to:
 *   1. Call setBudget(LOADED_BUDGET) via Effect 1.
 *   2. Effect 2 runs, sees skipNextSaveRef.current === true, flips it to false.
 * After this function returns the skip flag has been consumed and the next
 * user edit will trigger an auto-save.
 */
async function renderWithLoadedBudget() {
  useBudgets.mockReturnValue({
    loading: false,
    data: { budgets: [LOADED_BUDGET] },
    error: null,
  });
  useUpdateBudget.mockReturnValue([updateBudgetMock, {}]);

  let unmount;
  await act(async () => {
    const result = render(
      <BudgetProvider>
        <Harness />
      </BudgetProvider>
    );
    unmount = result.unmount;
  });
  return { unmount };
}

/**
 * Advances fake timers by `ms` milliseconds inside act() so that React state
 * updates triggered by timer callbacks are properly flushed. Also drains the
 * microtask queue so async mutation chains (setSaveStatus calls) are settled.
 */
async function advanceTime(ms) {
  await act(async () => {
    vi.advanceTimersByTime(ms);
  });
  // Drain any remaining microtasks / resolved promises.
  await act(async () => {});
}

// ---------------------------------------------------------------------------
// Setup / teardown
// ---------------------------------------------------------------------------

beforeEach(() => {
  vi.useFakeTimers();
  updateBudgetMock.mockReset();
  updateBudgetMock.mockResolvedValue({});
});

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

// ---------------------------------------------------------------------------
// REQ-001 — Auto-save debounce (TC-001 to TC-004)
// ---------------------------------------------------------------------------

describe('REQ-001 — Auto-save debounce', () => {
  it('TC-001: fires mutation after 1000ms', async () => {
    await renderWithLoadedBudget();

    await act(async () => {
      screen.getByTestId('edit-btn').click();
    });

    expect(updateBudgetMock).not.toHaveBeenCalled();

    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(1);
  });

  it('TC-002: does not fire before 1000ms', async () => {
    await renderWithLoadedBudget();

    await act(async () => {
      screen.getByTestId('edit-btn').click();
    });

    await advanceTime(999);

    expect(updateBudgetMock).toHaveBeenCalledTimes(0);
  });

  it('TC-003: rapid edits produce one mutation call', async () => {
    await renderWithLoadedBudget();

    // All five clicks fire synchronously inside a single act() — React 18
    // automatic batching collapses them into one re-render, producing a
    // single effect run and therefore a single debounce instance.
    await act(async () => {
      for (let i = 0; i < 5; i++) {
        screen.getByTestId('edit-btn').click();
      }
    });

    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(1);
  });

  it('TC-004: two quiet periods produce two mutation calls', async () => {
    await renderWithLoadedBudget();

    await act(async () => {
      screen.getByTestId('edit-btn').click();
    });
    await advanceTime(1000);
    expect(updateBudgetMock).toHaveBeenCalledTimes(1);

    await act(async () => {
      screen.getByTestId('edit-btn').click();
    });
    await advanceTime(1000);
    expect(updateBudgetMock).toHaveBeenCalledTimes(2);
  });
});

// ---------------------------------------------------------------------------
// REQ-002 — Skip initial load (TC-005 to TC-007)
// ---------------------------------------------------------------------------

describe('REQ-002 — Skip initial load', () => {
  it('TC-005: does not fire mutation on initial server data population', async () => {
    // renderWithLoadedBudget triggers the server-load path and consumes the
    // skip flag. No user edit follows — the server load itself must not save.
    await renderWithLoadedBudget();

    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(0);
  });

  it('TC-006: fires on first user edit after server load', async () => {
    await renderWithLoadedBudget();

    // Skip flag already consumed — this edit should trigger a save.
    await act(async () => {
      screen.getByTestId('edit-btn').click();
    });
    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(1);
  });

  it('TC-007: skip flag consumed only once — two subsequent edits each fire exactly once per 1000ms window', async () => {
    await renderWithLoadedBudget();

    await act(async () => {
      screen.getByTestId('edit-btn').click();
    });
    await advanceTime(1000);
    expect(updateBudgetMock).toHaveBeenCalledTimes(1);

    await act(async () => {
      screen.getByTestId('edit-btn').click();
    });
    await advanceTime(1000);
    expect(updateBudgetMock).toHaveBeenCalledTimes(2);
  });
});

// ---------------------------------------------------------------------------
// REQ-005 — No id = no save (TC-018 to TC-020)
// ---------------------------------------------------------------------------

describe('REQ-005 — No id = no save', () => {
  /**
   * Renders BudgetProvider with a fully loaded budget (so context is
   * accessible), consumes the skip flag, then exposes a button that calls
   * setBudget with the given no-id budget. This lets us exercise the
   * `!budget.id` guard in Effect 2 from within the context.
   */
  async function renderAndSetIdlessBudget(budget) {
    useBudgets.mockReturnValue({
      loading: false,
      data: { budgets: [LOADED_BUDGET] },
      error: null,
    });
    useUpdateBudget.mockReturnValue([updateBudgetMock, {}]);

    function IdlessHarness() {
      const { setBudget } = useBudget();
      return (
        <button data-testid="set-btn" onClick={() => setBudget(budget)}>
          Set
        </button>
      );
    }

    await act(async () => {
      render(
        <BudgetProvider>
          <IdlessHarness />
        </BudgetProvider>
      );
    });
    // Skip flag is now consumed (server load ran Effect 2 with id-bearing budget).
  }

  it('TC-018: no mutation when id is undefined', async () => {
    await renderAndSetIdlessBudget({ name: 'X', incomes: [], expenses: [] });

    await act(async () => {
      screen.getByTestId('set-btn').click();
    });
    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(0);
  });

  it('TC-019: no mutation when id is null', async () => {
    await renderAndSetIdlessBudget({ id: null, name: 'X', incomes: [], expenses: [] });

    await act(async () => {
      screen.getByTestId('set-btn').click();
    });
    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(0);
  });

  it('TC-020: no mutation when id is empty string', async () => {
    await renderAndSetIdlessBudget({ id: '', name: 'X', incomes: [], expenses: [] });

    await act(async () => {
      screen.getByTestId('set-btn').click();
    });
    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(0);
  });
});

// ---------------------------------------------------------------------------
// Edge cases (TC-021 to TC-024)
// ---------------------------------------------------------------------------

describe('Edge cases', () => {
  it('TC-021: pending debounce is cancelled on unmount — mutation never fires', async () => {
    const { unmount } = await renderWithLoadedBudget();

    // Trigger edit; skip flag already consumed by the server-load render.
    await act(async () => {
      screen.getByTestId('edit-btn').click();
    });

    // Unmount before the 1000ms window expires.
    await act(async () => {
      unmount();
    });

    // Advance past debounce window — cleanup in the effect should have
    // cancelled the pending debounce.
    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(0);
  });

  it('TC-022: no unmounted-component React warning after unmount then timer advance', async () => {
    const errorSpy = vi.spyOn(console, 'error');

    const { unmount } = await renderWithLoadedBudget();

    await act(async () => {
      screen.getByTestId('edit-btn').click();
    });

    await act(async () => {
      unmount();
    });

    await advanceTime(1000);

    const suspectMessages = errorSpy.mock.calls.filter((args) =>
      args.some(
        (a) =>
          typeof a === 'string' &&
          (a.includes('unmounted') || a.includes('memory leak') || a.includes('Warning:'))
      )
    );
    expect(suspectMessages).toHaveLength(0);

    errorSpy.mockRestore();
  });

  it('TC-023: setting an income amount to zero triggers a save', async () => {
    useBudgets.mockReturnValue({
      loading: false,
      data: { budgets: [LOADED_BUDGET] },
      error: null,
    });
    useUpdateBudget.mockReturnValue([updateBudgetMock, {}]);

    function ZeroAmountHarness() {
      const { setBudget } = useBudget();
      return (
        <button
          data-testid="zero-btn"
          onClick={() =>
            setBudget({
              id: '1',
              name: 'Test Budget',
              incomes: [{ source: 'Work', amount: 0 }],
              expenses: [],
            })
          }
        >
          Zero
        </button>
      );
    }

    await act(async () => {
      render(
        <BudgetProvider>
          <ZeroAmountHarness />
        </BudgetProvider>
      );
    });

    // Skip flag consumed by server-load effect; this click is a real user edit.
    await act(async () => {
      screen.getByTestId('zero-btn').click();
    });

    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(1);
  });

  it('TC-024: setting an empty budget name triggers a save', async () => {
    useBudgets.mockReturnValue({
      loading: false,
      data: { budgets: [LOADED_BUDGET] },
      error: null,
    });
    useUpdateBudget.mockReturnValue([updateBudgetMock, {}]);

    function EmptyNameHarness() {
      const { setBudget } = useBudget();
      return (
        <button
          data-testid="empty-name-btn"
          onClick={() =>
            setBudget({ id: '1', name: '', incomes: [], expenses: [] })
          }
        >
          Clear Name
        </button>
      );
    }

    await act(async () => {
      render(
        <BudgetProvider>
          <EmptyNameHarness />
        </BudgetProvider>
      );
    });

    // Skip flag consumed by server-load effect.
    await act(async () => {
      screen.getByTestId('empty-name-btn').click();
    });

    await advanceTime(1000);

    expect(updateBudgetMock).toHaveBeenCalledTimes(1);
  });
});
