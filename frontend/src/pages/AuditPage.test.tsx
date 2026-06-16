import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { LangProvider } from '../i18n';
import { api } from '../api';
import AuditPage from './AuditPage';

vi.mock('../api', () => ({
  api: {
    listAudits: vi.fn().mockResolvedValue([]),
    getCostStats: vi.fn().mockResolvedValue({ total_usd: 0, by_model: {}, total_input_tokens: 0, total_output_tokens: 0, audit_count: 0 }),
    createAudit: vi.fn().mockResolvedValue({ id: 'a1' }),
    deleteAudit: vi.fn(),
  },
}));

function renderPage() {
  return render(
    <MemoryRouter>
      <LangProvider>
        <AuditPage />
      </LangProvider>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  localStorage.clear();
  vi.clearAllMocks();
});

describe('AuditPage', () => {
  it('submits a new static-snapshot audit for the entered repo', async () => {
    const user = userEvent.setup();
    renderPage();

    const repoInput = await screen.findByPlaceholderText('e.g. directus/directus');
    await user.type(repoInput, 'torvalds/linux');

    const runButton = screen.getByRole('button', { name: /run audit/i });
    await user.click(runButton);

    await waitFor(() => {
      expect(api.createAudit).toHaveBeenCalledWith(
        expect.objectContaining({ repo: 'torvalds/linux' }),
      );
    });
  });
});
