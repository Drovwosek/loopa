import { describe, it, expect, vi, beforeEach } from 'vitest';
import { configureStore } from '@reduxjs/toolkit';
import taskReducer, { loadTask, clearTask } from './taskSlice';
import * as api from '../api';

// Mock the API module
vi.mock('../api');

describe('taskSlice', () => {
  const createTestStore = () =>
    configureStore({
      reducer: {
        task: taskReducer,
      },
    });

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const store = createTestStore();
      const state = store.getState().task;

      expect(state.loading).toBe(false);
      expect(state.current).toBeUndefined();
      expect(state.error).toBeUndefined();
    });
  });

  describe('clearTask', () => {
    it('should clear current task and error', () => {
      const store = createTestStore();

      // First set some state
      store.dispatch(clearTask());

      const state = store.getState().task;
      expect(state.current).toBeUndefined();
      expect(state.error).toBeUndefined();
    });
  });

  describe('loadTask', () => {
    it('should set loading to true when pending', async () => {
      const mockTask = {
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      };

      vi.mocked(api.fetchTask).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockTask), 100))
      );

      const store = createTestStore();
      const promise = store.dispatch(loadTask('task-123'));

      // Check loading state
      expect(store.getState().task.loading).toBe(true);

      await promise;
    });

    it('should set current task when fulfilled', async () => {
      const mockTask = {
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      };

      vi.mocked(api.fetchTask).mockResolvedValue(mockTask);

      const store = createTestStore();
      await store.dispatch(loadTask('task-123'));

      const state = store.getState().task;
      expect(state.loading).toBe(false);
      expect(state.current).toEqual(mockTask);
      expect(state.error).toBeUndefined();
    });

    it('should set error when rejected', async () => {
      vi.mocked(api.fetchTask).mockRejectedValue(new Error('Network error'));

      const store = createTestStore();
      await store.dispatch(loadTask('task-123'));

      const state = store.getState().task;
      expect(state.loading).toBe(false);
      expect(state.current).toBeUndefined();
      expect(state.error).toBe('Network error');
    });

    it('should have default error message when error has no message', async () => {
      vi.mocked(api.fetchTask).mockRejectedValue({});

      const store = createTestStore();
      await store.dispatch(loadTask('task-123'));

      const state = store.getState().task;
      expect(state.error).toBe('Failed to load task');
    });
  });
});
