import { describe, it, expect, vi, beforeEach } from 'vitest';
import { configureStore } from '@reduxjs/toolkit';
import historyReducer, { loadHistory } from './historySlice';
import * as api from '../api';

// Mock the API module
vi.mock('../api');

describe('historySlice', () => {
  const createTestStore = () =>
    configureStore({
      reducer: {
        history: historyReducer,
      },
    });

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const store = createTestStore();
      const state = store.getState().history;

      expect(state.items).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBeUndefined();
    });
  });

  describe('loadHistory', () => {
    it('should set loading to true when pending', async () => {
      const mockHistory = [
        {
          id: 'task-1',
          originalName: 'file1.mp3',
          status: 'готово',
          uploadedAt: '2024-01-01T00:00:00Z',
        },
      ];

      vi.mocked(api.fetchHistory).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockHistory), 100))
      );

      const store = createTestStore();
      const promise = store.dispatch(loadHistory());

      expect(store.getState().history.loading).toBe(true);

      await promise;
    });

    it('should set items when fulfilled', async () => {
      const mockHistory = [
        {
          id: 'task-1',
          originalName: 'file1.mp3',
          status: 'готово',
          uploadedAt: '2024-01-01T00:00:00Z',
        },
        {
          id: 'task-2',
          originalName: 'file2.wav',
          status: 'ожидает',
          uploadedAt: '2024-01-02T00:00:00Z',
        },
      ];

      vi.mocked(api.fetchHistory).mockResolvedValue(mockHistory);

      const store = createTestStore();
      await store.dispatch(loadHistory());

      const state = store.getState().history;
      expect(state.loading).toBe(false);
      expect(state.items).toEqual(mockHistory);
      expect(state.error).toBeUndefined();
    });

    it('should set error when rejected', async () => {
      vi.mocked(api.fetchHistory).mockRejectedValue(new Error('Failed to fetch'));

      const store = createTestStore();
      await store.dispatch(loadHistory());

      const state = store.getState().history;
      expect(state.loading).toBe(false);
      expect(state.items).toEqual([]);
      expect(state.error).toBe('Failed to fetch');
    });

    it('should have default error message when error has no message', async () => {
      vi.mocked(api.fetchHistory).mockRejectedValue({});

      const store = createTestStore();
      await store.dispatch(loadHistory());

      const state = store.getState().history;
      expect(state.error).toBe('Failed to load history');
    });

    it('should handle empty history', async () => {
      vi.mocked(api.fetchHistory).mockResolvedValue([]);

      const store = createTestStore();
      await store.dispatch(loadHistory());

      const state = store.getState().history;
      expect(state.loading).toBe(false);
      expect(state.items).toEqual([]);
      expect(state.error).toBeUndefined();
    });
  });
});
