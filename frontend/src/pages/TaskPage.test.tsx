import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Provider } from 'react-redux';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { configureStore } from '@reduxjs/toolkit';
import TaskPage from './TaskPage';
import historyReducer from '../store/historySlice';
import taskReducer from '../store/taskSlice';
import projectReducer from '../store/projectSlice';
import * as api from '../api';

vi.mock('../api');

// Mock complex child components that depend on browser APIs (WaveSurfer, etc.)
vi.mock('../components/player/AudioPlayer', () => ({
  default: ({ onTimeUpdate }: { audioUrl: string; onTimeUpdate: (ms: number) => void }) => (
    <div data-testid="audio-player">AudioPlayer mock</div>
  ),
}));

vi.mock('../components/editor/TranscriptionEditor', () => ({
  default: () => <div data-testid="transcription-editor">TranscriptionEditor mock</div>,
}));

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

const renderTaskPage = (taskId = 'task-123') => {
  const store = createTestStore();
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
    it('shows loading spinner initially', () => {
      vi.mocked(api.fetchTask).mockImplementation(
        () => new Promise(() => {}) // never resolves
      );

      renderTaskPage();
      // Ant Design Spin uses aria-busy="true"
      expect(document.querySelector('.ant-spin')).toBeInTheDocument();
    });

    it('displays task name and status after loading', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.fetchSegments).mockResolvedValue([]);

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('test.mp3')).toBeInTheDocument();
      });
    });

    it('displays error message when task has error', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'ошибка',
        originalName: 'test.mp3',
        errorMessage: 'Recognition failed',
        createdAt: '2024-01-01T00:00:00Z',
      });

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('Recognition failed')).toBeInTheDocument();
      });
    });

    it('shows AudioPlayer and TranscriptionEditor for completed task with segments', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.fetchSegments).mockResolvedValue([
        {
          id: 'seg-1',
          speakerId: 'spk-1',
          speakerName: 'Спикер 1',
          startTime: 0,
          endTime: 5000,
          text: 'Hello world',
          hasFillers: false,
          isCorrected: false,
        },
      ]);

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByTestId('audio-player')).toBeInTheDocument();
      });
      expect(screen.getByTestId('transcription-editor')).toBeInTheDocument();
    });

    it('shows transcript text fallback when no segments', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world transcript',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.fetchSegments).mockResolvedValue([]);

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('Hello world transcript')).toBeInTheDocument();
      });
    });

    it('shows error alert when task loading fails', async () => {
      vi.mocked(api.fetchTask).mockRejectedValue(new Error('Ошибка загрузки'));

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('Ошибка загрузки')).toBeInTheDocument();
      });
    });
  });

  describe('Actions', () => {
    it('Назад button navigates to home', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.fetchSegments).mockResolvedValue([]);

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('Назад')).toBeInTheDocument();
      });

      const backButton = screen.getByText('Назад');
      await userEvent.click(backButton);

      expect(mockNavigate).toHaveBeenCalledWith('/');
    });

    it('copy button copies transcript to clipboard', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.fetchSegments).mockResolvedValue([]);

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('Копировать текст')).toBeInTheDocument();
      });

      const copyButton = screen.getByText('Копировать текст');
      await userEvent.click(copyButton);

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith('Hello world');
    });

    it('Скачать TXT button calls downloadExport', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.fetchSegments).mockResolvedValue([]);
      vi.mocked(api.downloadExport).mockResolvedValue(undefined);

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('Скачать TXT')).toBeInTheDocument();
      });

      const downloadTxtButton = screen.getByText('Скачать TXT');
      await userEvent.click(downloadTxtButton);

      await waitFor(() => {
        expect(api.downloadExport).toHaveBeenCalledWith('task-123', 'txt');
      });
    });

    it('Скачать DOCX button calls downloadExport', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      });
      vi.mocked(api.fetchSegments).mockResolvedValue([]);
      vi.mocked(api.downloadExport).mockResolvedValue(undefined);

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('Скачать DOCX')).toBeInTheDocument();
      });

      const downloadDocxButton = screen.getByText('Скачать DOCX');
      await userEvent.click(downloadDocxButton);

      await waitFor(() => {
        expect(api.downloadExport).toHaveBeenCalledWith('task-123', 'docx');
      });
    });
  });

  describe('Processing state', () => {
    it('shows processing message for in-progress task', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'в процессе',
        originalName: 'test.mp3',
        createdAt: '2024-01-01T00:00:00Z',
      });

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('Идёт обработка...')).toBeInTheDocument();
      });
    });

    it('shows processing message for pending task', async () => {
      vi.mocked(api.fetchTask).mockResolvedValue({
        id: 'task-123',
        status: 'ожидает',
        originalName: 'test.mp3',
        createdAt: '2024-01-01T00:00:00Z',
      });

      renderTaskPage();

      await waitFor(() => {
        expect(screen.getByText('Идёт обработка...')).toBeInTheDocument();
      });
    });
  });
});
