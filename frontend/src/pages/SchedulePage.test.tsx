import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LangProvider } from '../i18n';
import { api } from '../api';
import SchedulePage from './SchedulePage';
import type { Schedule } from '../api';

const sampleSchedule: Schedule = {
  id: 's1',
  repo: 'torvalds/linux',
  interval_h: 168,
  provider: '',
  model: '',
  enabled: true,
  time_window_start: -1,
  time_window_end: -1,
  cli_fallback: false,
  auto_detected: false,
  last_run_at: null,
  next_run_at: new Date().toISOString(),
  created_at: new Date().toISOString(),
};

vi.mock('../api', () => ({
  api: {
    listSchedules: vi.fn().mockResolvedValue([]),
    createSchedule: vi.fn(),
    updateSchedule: vi.fn(),
    deleteSchedule: vi.fn(),
    triggerSchedule: vi.fn(),
    pauseAllSchedules: vi.fn().mockResolvedValue(undefined),
  },
}));

function renderPage() {
  return render(
    <LangProvider>
      <SchedulePage />
    </LangProvider>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.spyOn(window, 'confirm').mockReturnValue(true);
  (api.listSchedules as ReturnType<typeof vi.fn>).mockResolvedValue([sampleSchedule]);
});

describe('SchedulePage', () => {
  it('pauses all active schedules in one click', async () => {
    const user = userEvent.setup();
    renderPage();

    const pauseAllButton = await screen.findByRole('button', { name: /pause all/i });
    await user.click(pauseAllButton);

    expect(window.confirm).toHaveBeenCalled();
    await waitFor(() => {
      expect(api.pauseAllSchedules).toHaveBeenCalledTimes(1);
    });
    await waitFor(() => {
      expect(api.listSchedules).toHaveBeenCalledTimes(2); // initial load + reload after pause
    });
  });
});
