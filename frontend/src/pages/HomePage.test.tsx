import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';
import { configureStore } from '@reduxjs/toolkit';
import HomePage from './HomePage';
import historyReducer from '../store/historySlice';
import taskReducer from '../store/taskSlice';
import projectReducer from '../store/projectSlice';
import * as api from '../api';

vi.mock('../api');

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const createTestStore = (preloadedState = {}) =>
  configureStore({
    reducer: {
      history: historyReducer,
      task: taskReducer,
      projects: projectReducer,
    },
    preloadedState,
  });

const renderHomePage = (preloadedState = {}) => {
  const store = createTestStore(preloadedState);
  return {
    store,
    ...render(
      <Provider store={store}>
        <MemoryRouter>
          <HomePage />
        </MemoryRouter>
      </Provider>
    ),
  };
};

describe('HomePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.fetchHistory).mockResolvedValue([]);
    vi.mocked(api.fetchProjects).mockResolvedValue([]);
  });

  describe('Upload section', () => {
    it('renders upload section', () => {
      renderHomePage();
      expect(screen.getByText('Загрузка медиафайла')).toBeInTheDocument();
      expect(screen.getByText(/MP3, WAV, MP4, MOV/)).toBeInTheDocument();
    });

    it('navigates to task page after successful upload', async () => {
      vi.mocked(api.uploadFile).mockResolvedValue('task-123');

      renderHomePage();

      const file = new File(['test'], 'test.mp3', { type: 'audio/mpeg' });
      const input = document.querySelector('input[type="file"]') as HTMLInputElement;

      if (input) {
        await userEvent.upload(input, file);

        await waitFor(() => {
          expect(screen.getByText('Начать транскрибацию')).toBeInTheDocument();
        });

        const uploadButton = screen.getByText('Начать транскрибацию');
        await userEvent.click(uploadButton);

        await waitFor(() => {
          expect(mockNavigate).toHaveBeenCalledWith('/tasks/task-123');
        });
      }
    });

    it('shows error on upload failure', async () => {
      vi.mocked(api.uploadFile).mockRejectedValue(new Error('Ошибка загрузки'));

      renderHomePage();

      const file = new File(['test'], 'test.mp3', { type: 'audio/mpeg' });
      const input = document.querySelector('input[type="file"]') as HTMLInputElement;

      if (input) {
        await userEvent.upload(input, file);

        await waitFor(() => {
          expect(screen.getByText('Начать транскрибацию')).toBeInTheDocument();
        });

        const uploadButton = screen.getByText('Начать транскрибацию');
        await userEvent.click(uploadButton);

        await waitFor(() => {
          expect(screen.getByText('Ошибка загрузки')).toBeInTheDocument();
        });
      }
    });
  });

  describe('History section', () => {
    it('shows empty message when no history', async () => {
      vi.mocked(api.fetchHistory).mockResolvedValue([]);

      renderHomePage();

      await waitFor(() => {
        expect(screen.getByText('Нет загрузок')).toBeInTheDocument();
      });
    });

    it('renders history items after loading', async () => {
      const historyItems = [
        { id: 'task-1', originalName: 'file1.mp3', status: 'готово', uploadedAt: '2024-01-01T00:00:00Z' },
        { id: 'task-2', originalName: 'file2.wav', status: 'ожидает', uploadedAt: '2024-01-02T00:00:00Z' },
      ];

      vi.mocked(api.fetchHistory).mockResolvedValue(historyItems);

      renderHomePage();

      await waitFor(() => {
        expect(screen.getByText('file1.mp3')).toBeInTheDocument();
      });

      expect(screen.getByText('file2.wav')).toBeInTheDocument();
    });

    it('navigates to task on Открыть button click', async () => {
      const historyItems = [
        { id: 'task-1', originalName: 'file1.mp3', status: 'готово', uploadedAt: '2024-01-01T00:00:00Z' },
      ];

      vi.mocked(api.fetchHistory).mockResolvedValue(historyItems);

      renderHomePage();

      await waitFor(() => {
        expect(screen.getByText('file1.mp3')).toBeInTheDocument();
      });

      const openButton = screen.getByText('Открыть');
      await userEvent.click(openButton);

      expect(mockNavigate).toHaveBeenCalledWith('/tasks/task-1');
    });

    it('calls deleteTask on confirm', async () => {
      vi.mocked(api.deleteTask).mockResolvedValue(undefined);

      const historyItems = [
        { id: 'task-1', originalName: 'file1.mp3', status: 'готово', uploadedAt: '2024-01-01T00:00:00Z' },
      ];

      vi.mocked(api.fetchHistory).mockResolvedValue(historyItems);

      renderHomePage();

      await waitFor(() => {
        expect(screen.getByText('file1.mp3')).toBeInTheDocument();
      });

      const deleteButton = screen.getByText('Удалить');
      await userEvent.click(deleteButton);

      // Popconfirm renders a tooltip with confirm/cancel buttons
      // In Ant Design v6, find and click the confirm button inside the popover
      await waitFor(() => {
        const popconfirmButtons = document.querySelectorAll('.ant-popconfirm .ant-btn-primary, .ant-popover .ant-btn-primary');
        expect(popconfirmButtons.length).toBeGreaterThan(0);
      });
      const confirmBtn = document.querySelector('.ant-popconfirm .ant-btn-primary, .ant-popover .ant-btn-primary') as HTMLElement;
      await userEvent.click(confirmBtn);

      await waitFor(() => {
        expect(api.deleteTask).toHaveBeenCalledWith('task-1');
      });
    });
  });
});
