import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';
import { configureStore } from '@reduxjs/toolkit';
import HomePage from './HomePage';
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
  });

  describe('Upload section', () => {
    it('renders upload section', () => {
      renderHomePage();
      expect(screen.getByText(/upload media/i)).toBeInTheDocument();
      expect(screen.getByText(/mp3, wav, mp4, mov/i)).toBeInTheDocument();
    });

    it('shows no file selected initially', () => {
      renderHomePage();
      expect(screen.getByText(/no file selected/i)).toBeInTheDocument();
    });

    it('shows error when trying to upload without file', async () => {
      renderHomePage();
      const uploadButton = screen.getByRole('button', { name: /upload/i });
      
      await userEvent.click(uploadButton);
      
      expect(screen.getByText(/select a file first/i)).toBeInTheDocument();
    });

    it('navigates to task page after successful upload', async () => {
      vi.mocked(api.uploadFile).mockResolvedValue('task-123');
      
      renderHomePage();
      
      const file = new File(['test'], 'test.mp3', { type: 'audio/mpeg' });
      const input = document.querySelector('input[type="file"]') as HTMLInputElement;
      
      if (input) {
        Object.defineProperty(input, 'files', {
          value: [file],
          configurable: true,
        });
        fireEvent.change(input);
        
        const uploadButton = screen.getByRole('button', { name: /upload/i });
        await userEvent.click(uploadButton);
        
        await waitFor(() => {
          expect(mockNavigate).toHaveBeenCalledWith('/tasks/task-123');
        });
      }
    });

    it('shows error on upload failure', async () => {
      vi.mocked(api.uploadFile).mockRejectedValue(new Error('Upload failed'));
      
      renderHomePage();
      
      const file = new File(['test'], 'test.mp3', { type: 'audio/mpeg' });
      const input = document.querySelector('input[type="file"]') as HTMLInputElement;
      
      if (input) {
        Object.defineProperty(input, 'files', {
          value: [file],
          configurable: true,
        });
        fireEvent.change(input);
        
        const uploadButton = screen.getByRole('button', { name: /upload/i });
        await userEvent.click(uploadButton);
        
        await waitFor(() => {
          expect(screen.getByText(/upload failed/i)).toBeInTheDocument();
        });
      }
    });
  });

  describe('History section', () => {
    it('shows loading state while fetching history', async () => {
      // Mock a slow response
      vi.mocked(api.fetchHistory).mockImplementation(() => 
        new Promise(resolve => setTimeout(() => resolve([]), 100))
      );
      
      renderHomePage();
      
      expect(screen.getByText(/loading/i)).toBeInTheDocument();
    });

    it('shows no uploads message when history is empty', async () => {
      vi.mocked(api.fetchHistory).mockResolvedValue([]);
      
      renderHomePage();
      
      await waitFor(() => {
        expect(screen.getByText(/no uploads yet/i)).toBeInTheDocument();
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
      expect(screen.getByText('готово')).toBeInTheDocument();
      expect(screen.getByText('ожидает')).toBeInTheDocument();
    });

    it('navigates to task on Open button click', async () => {
      const historyItems = [
        { id: 'task-1', originalName: 'file1.mp3', status: 'готово', uploadedAt: '2024-01-01T00:00:00Z' },
      ];
      
      vi.mocked(api.fetchHistory).mockResolvedValue(historyItems);
      
      renderHomePage();
      
      await waitFor(() => {
        expect(screen.getByText('file1.mp3')).toBeInTheDocument();
      });
      
      const openButton = screen.getByRole('button', { name: /open/i });
      await userEvent.click(openButton);
      
      expect(mockNavigate).toHaveBeenCalledWith('/tasks/task-1');
    });

    it('deletes task on Delete button click', async () => {
      vi.mocked(api.deleteTask).mockResolvedValue(undefined);
      
      const historyItems = [
        { id: 'task-1', originalName: 'file1.mp3', status: 'готово', uploadedAt: '2024-01-01T00:00:00Z' },
      ];
      
      vi.mocked(api.fetchHistory).mockResolvedValue(historyItems);
      
      renderHomePage();
      
      await waitFor(() => {
        expect(screen.getByText('file1.mp3')).toBeInTheDocument();
      });
      
      const deleteButton = screen.getByRole('button', { name: /delete/i });
      await userEvent.click(deleteButton);
      
      await waitFor(() => {
        expect(api.deleteTask).toHaveBeenCalledWith('task-1');
      });
    });
  });

  describe('Drag and drop', () => {
    it('handles drag over event', () => {
      renderHomePage();
      const dropZone = screen.getByText(/upload media/i).closest('section');
      
      if (dropZone) {
        fireEvent.dragOver(dropZone);
        expect(dropZone).toHaveStyle({ border: expect.stringContaining('dashed') });
      }
    });
  });
});
