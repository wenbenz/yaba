/**
 * Integration tests for BudgetProvider + BudgetEditor combined.
 * Covers: TC-008–TC-017
 *
 * These tests render the real BudgetEditor inside BudgetProvider and assert
 * on the SaveStatusLabel text that appears in the DOM. Heavy child components
 * (Income, Expense) are stubbed to keep rendering lightweight and deterministic.
 */

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

// ---------------------------------------------------------------------------
// Mocks — must be declared before any import that transitively needs them.
// ---------------------------------------------------------------------------
const updateBudgetMock = vi.fn();

vi.mock('../../../api/graph', () => ({
  useBudgets: vi.fn(),
  useUpdateBudget: vi.fn(),
}));

vi.mock('../../../components/Loader', () => ({ default: () => <div data-testid="loader" /> }));
vi.mock('../CreateBudget', () => ({ default: () => <div data-testid="create-budget" /> }));

// Stub sub-components used inside BudgetEditor so they don't need context or
// their own deps (they still call useBudget, which IS available).
vi.mock('../Income', () => ({ default: ({ index }) => <div data-testid={`income-${index}`} /> }));
vi.mock('../Expense', () => ({ default: ({ index }) => <div data-testid={`expense-${index}`} /> }));

// ---------------------------------------------------------------------------
// Imports
// ---------------------------------------------------------------------------
import { useBudgets, useUpdateBudget } from '../../../api/graph';
import { BudgetProvider } from '../BudgetContext';
import BudgetEditor from '../BudgetEditor';

// ---------------------------------------------------------------------------
// Shared fixtures
// ---------------------------------------------------------------------------

const LOADED_BUDGET = {
  id: '1',
  name: 'Test Budget',
  incomes: [{ source: 'Work', amount: 2400 }],
  expenses: [
    { category: 'Rent', amount: 1000, isFixed: true, isSlack: false },
    { category: 'Misc', amount: 200, isFixed: false, isSlack: true },
  ],
};

/**
 * Renders BudgetProvider + BudgetEditor with the given useBudgets mock value
 * and the given updateBudgetMock behaviour.
 *
 * Returns { unmount } and waits for all effects (including the skip-flag
 * consumption on server load) to settle before returning.
 */
async function renderEditor({ budgetData = { budgets: [LOADED_BUDGET] } } = {}) {
  useBudgets.mockReturnValue({ loading: false, data: budgetData, error: null });
  useUpdateBudget.mockReturnValue([updateBudgetMock, {}]);

  let unmount;
  await act(async () => {
    const result = render(
      <BudgetProvider>
        <BudgetEditor />
      </BudgetProvider>
    );
    unmount = result.unmount;
  });
  return { unmount };
}

/**
 * Advances fake timers inside act() to flush both timer callbacks and any
 * resulting async state-update chains.
 */
async function advanceTime(ms) {
  await act(async () => {
    vi.advanceTimersByTime(ms);
  });
  await act(async () => {});
}

// Counter ensures each triggerUserEdit call uses a distinct value so React's
// controlled-input diffing recognises a real change and fires onChange.
let editCounter = 0;

/**
 * Simulates a user edit by setting the Budget Name field to a unique value,
 * which calls setBudget via the onChange handler in BudgetEditor.
 * Uses native DOM events because vi.useFakeTimers interferes with userEvent.
 */
async function triggerUserEdit() {
  editCounter += 1;
  const nameInput = screen.getByLabelText('Budget Name');
  await act(async () => {
    const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
      window.HTMLInputElement.prototype,
      'value'
    ).set;
    nativeInputValueSetter.call(nameInput, `Edited Budget ${editCounter}`);
    nameInput.dispatchEvent(new Event('input', { bubbles: true }));
    nameInput.dispatchEvent(new Event('change', { bubbles: true }));
  });
}

// ---------------------------------------------------------------------------
// Setup / teardown
// ---------------------------------------------------------------------------

beforeEach(() => {
  vi.useFakeTimers();
  editCounter = 0;
  updateBudgetMock.mockReset();
  updateBudgetMock.mockResolvedValue({});
});

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

// ---------------------------------------------------------------------------
// REQ-003 — saveStatus label transitions (TC-008 to TC-016)
// ---------------------------------------------------------------------------

describe('REQ-003 — SaveStatusLabel DOM transitions', () => {
  it('TC-008: no status label rendered on idle (no edit triggered)', async () => {
    await renderEditor();

    expect(screen.queryByText('Saving…')).not.toBeInTheDocument();
    expect(screen.queryByText('Saved')).not.toBeInTheDocument();
    expect(screen.queryByText('Save failed')).not.toBeInTheDocument();
  });

  it('TC-009: shows "Saving…" while mutation is in-flight', async () => {
    // Return a never-resolving promise so we can assert the intermediate state.
    updateBudgetMock.mockReturnValue(new Promise(() => {}));

    await renderEditor();

    await triggerUserEdit();
    await advanceTime(1000);

    expect(screen.getByText('Saving…')).toBeInTheDocument();
  });

  it('TC-010: shows "Saved" after mutation resolves, "Saving…" absent', async () => {
    updateBudgetMock.mockResolvedValue({});

    await renderEditor();

    await triggerUserEdit();
    await advanceTime(1000);

    expect(screen.getByText('Saved')).toBeInTheDocument();
    expect(screen.queryByText('Saving…')).not.toBeInTheDocument();
  });

  it('TC-011: "Saved" auto-clears to idle after 2000ms', async () => {
    updateBudgetMock.mockResolvedValue({});

    await renderEditor();

    await triggerUserEdit();
    await advanceTime(1000); // debounce fires + mutation resolves → "Saved"

    expect(screen.getByText('Saved')).toBeInTheDocument();

    await advanceTime(2000); // clear-timer fires → idle

    expect(screen.queryByText('Saved')).not.toBeInTheDocument();
  });

  it('TC-012: "Saved" still present at 1999ms after success', async () => {
    updateBudgetMock.mockResolvedValue({});

    await renderEditor();

    await triggerUserEdit();
    await advanceTime(1000);

    expect(screen.getByText('Saved')).toBeInTheDocument();

    await advanceTime(1999); // one ms short of the clear-timer

    expect(screen.getByText('Saved')).toBeInTheDocument();
  });

  it('TC-013: new edit before "Saved" clears shows "Saving…" and removes "Saved"', async () => {
    updateBudgetMock.mockResolvedValue({});

    await renderEditor();

    // First save cycle — resolves and sets "Saved".
    await triggerUserEdit();
    await advanceTime(1000);
    expect(screen.getByText('Saved')).toBeInTheDocument();

    // Advance 500ms into the 2000ms clear window.
    await advanceTime(500);
    expect(screen.getByText('Saved')).toBeInTheDocument();

    // Second edit: start a new in-flight save that won't resolve.
    updateBudgetMock.mockReturnValue(new Promise(() => {}));
    await triggerUserEdit();
    await advanceTime(1000); // second debounce fires

    expect(screen.getByText('Saving…')).toBeInTheDocument();
    expect(screen.queryByText('Saved')).not.toBeInTheDocument();
  });

  it('TC-014: shows "Save failed" when mutation rejects', async () => {
    updateBudgetMock.mockRejectedValue(new Error('network error'));

    await renderEditor();

    await triggerUserEdit();
    await advanceTime(1000);

    expect(screen.getByText('Save failed')).toBeInTheDocument();
  });

  it('TC-015: "Save failed" persists after 2000ms', async () => {
    updateBudgetMock.mockRejectedValue(new Error('network error'));

    await renderEditor();

    await triggerUserEdit();
    await advanceTime(1000);
    expect(screen.getByText('Save failed')).toBeInTheDocument();

    await advanceTime(2000);

    expect(screen.getByText('Save failed')).toBeInTheDocument();
  });

  it('TC-016: "Save failed" clears when subsequent save succeeds', async () => {
    updateBudgetMock.mockRejectedValue(new Error('network error'));

    await renderEditor();

    // First save — fails.
    await triggerUserEdit();
    await advanceTime(1000);
    expect(screen.getByText('Save failed')).toBeInTheDocument();

    // Second save — succeeds.
    updateBudgetMock.mockResolvedValue({});
    await triggerUserEdit();
    await advanceTime(1000);

    expect(screen.getByText('Saved')).toBeInTheDocument();
    expect(screen.queryByText('Save failed')).not.toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// REQ-004 — No Save button (TC-017)
// ---------------------------------------------------------------------------

describe('REQ-004 — No Save button', () => {
  it('TC-017: no button with accessible name matching /save/i is rendered', async () => {
    await renderEditor();

    expect(
      screen.queryByRole('button', { name: /save/i })
    ).not.toBeInTheDocument();
  });
});
