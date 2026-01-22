import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Provider } from 'react-redux';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { configureStore } from '@reduxjs/toolkit';
import TaskPage from './TaskPage';
import historyReducer from '../store/historySlice';
import taskReducer from '../store/taskSlice';
import * as api from '../api';

// Mock the API module
vi.mock('../api');

// Mock react-router-dom navigate
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
    },
    preloadedState,
  });

const renderTaskPage = (taskId = 'task-123', preloadedState = {}) => {
  const store = createTestStore(preloadedState);
  return {
    store,
    ...render(
      <Provider store={store}>
        <MemoryRouter initialEntries={[`/tasks/${taskId}`]}>
          <Routes>
            <Route path="/tasks/:id" element={<TaskPage />} />
          </Routes>
        </MemoryRouter>
      </Provider>
    ),
  };
};

describe('TaskPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading and display', () => {
    it('shows loading message initially', async () => {
      vi.mocked(api.fetchTask).mockImplementation(() =>
        new Promise(resolve => setTimeout(() => resolve({
          id: 'task-123',
          status: 'готово',
          originalName: 'test.mp3',
          transcriptText: 'Hello world',
          createdAt: '2024-01-01T00:00:00Z',
        }), 100))
      );

      renderTaskPage('task-123');
      
      expect(screen.getByText(/loading/i)).toBeInTheDocument();
    });

    it('displays task details after loading', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });

      renderTaskPage('task-123');

      await waitFor(() => {
        expect(screen.getByText(/task details/i)).toBeInTheDocument();
      });
      
      expect(screen.getByText('test.mp3')).toBeInTheDocument();
      expect(screen.getByText('готово')).toBeInTheDocument();
    });

    it('displays transcript when available', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world transcript',
        createdAt: '2024-01-01T00:00:00Z',
      });

      renderTaskPage('task-123');

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: /transcript/i })).toBeInTheDocument();
      });
      
      expect(screen.getByDisplayValue('Hello world transcript')).toBeInTheDocument();
    });

    it('displays error message when task has error', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'ошибка',
        originalName: 'test.mp3',
        errorMessage: 'Recognition failed',
        createdAt: '2024-01-01T00:00:00Z',
      });

      renderTaskPage('task-123');

      await waitFor(() => {
        expect(screen.getByText('Recognition failed')).toBeInTheDocument();
      });
    });

    it('shows task not found after failed load', async () => {
      vi.mocked(api.fetchTask).mockRejectedValue(new Error('Not found'));

      renderTaskPage('task-123');

      await waitFor(() => {
        expect(screen.getByText(/task not found/i)).toBeInTheDocument();
      });
    });
  });

  describe('Actions', () => {
    it('back button navigates to home', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });

      renderTaskPage('task-123');

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /back/i })).toBeInTheDocument();
      });

      const backButton = screen.getByRole('button', { name: /back/i });
      await userEvent.click(backButton);

      expect(mockNavigate).toHaveBeenCalledWith('/');
    });

    it('copy all button copies transcript to clipboard', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });

      renderTaskPage('task-123');

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /copy all/i })).toBeInTheDocument();
      });

      const copyButton = screen.getByRole('button', { name: /copy all/i });
      await userEvent.click(copyButton);

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith('Hello world');
    });

    it('download TXT button calls downloadExport', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.downloadExport).mockResolvedValue(undefined);

      renderTaskPage('task-123');

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /download txt/i })).toBeInTheDocument();
      });

      const downloadTxtButton = screen.getByRole('button', { name: /download txt/i });
      await userEvent.click(downloadTxtButton);

      await waitFor(() => {
        expect(api.downloadExport).toHaveBeenCalledWith('task-123', 'txt');
      });
    });

    it('download DOCX button calls downloadExport', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.downloadExport).mockResolvedValue(undefined);

      renderTaskPage('task-123');

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /download docx/i })).toBeInTheDocument();
      });

      const downloadDocxButton = screen.getByRole('button', { name: /download docx/i });
      await userEvent.click(downloadDocxButton);

      await waitFor(() => {
        expect(api.downloadExport).toHaveBeenCalledWith('task-123', 'docx');
      });
    });

    it('shows error on download failure', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.downloadExport).mockRejectedValue(new Error('Export failed'));

      renderTaskPage('task-123');

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /download txt/i })).toBeInTheDocument();
      });

      const downloadTxtButton = screen.getByRole('button', { name: /download txt/i });
      await userEvent.click(downloadTxtButton);

      await waitFor(() => {
        expect(screen.getByText(/export failed/i)).toBeInTheDocument();
      });
    });
  });

  describe('Task polling', () => {
    it('does not poll for completed tasks', async () => {
      vi.useFakeTimers();
      
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });

      renderTaskPage('task-123');
      
      // Wait for initial load
      await vi.runAllTimersAsync();
      
      // Clear mocks after initial load
      vi.mocked(api.fetchTask).mockClear();
      
      // Advance timer
      vi.advanceTimersByTime(5000);
      await vi.runAllTimersAsync();
      
      // Should not have called fetchTask again since task is completed
      expect(api.fetchTask).not.toHaveBeenCalled();
      
      vi.useRealTimers();
    });
  });
});
