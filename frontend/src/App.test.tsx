import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';
import { configureStore } from '@reduxjs/toolkit';
import App from './App';
import historyReducer from './store/historySlice';
import taskReducer from './store/taskSlice';

const createTestStore = (preloadedState = {}) =>
  configureStore({
    reducer: {
      history: historyReducer,
      task: taskReducer,
    },
    preloadedState,
  });

const renderApp = (initialEntries = ['/'], preloadedState = {}) => {
  const store = createTestStore(preloadedState);
  return render(
    <Provider store={store}>
      <MemoryRouter initialEntries={initialEntries}>
        <App />
      </MemoryRouter>
    </Provider>
  );
};

describe('App', () => {
  it('renders the header with Loopa title', () => {
    renderApp();
    expect(screen.getByRole('heading', { name: /loopa/i })).toBeInTheDocument();
  });

  it('renders HomePage at root route', () => {
    renderApp(['/']);
    expect(screen.getByText(/upload media/i)).toBeInTheDocument();
  });

  it('renders TaskPage at /tasks/:id route', () => {
    const preloadedState = {
      task: {
        loading: false,
        current: {
          id: 'task-123',
          status: 'готово',
          originalName: 'test.mp3',
          transcriptText: 'Hello world',
          createdAt: '2024-01-01T00:00:00Z',
        },
      },
      history: {
        items: [],
        loading: false,
      },
    };

    renderApp(['/tasks/task-123'], preloadedState);
    expect(screen.getByText(/task details/i)).toBeInTheDocument();
  });

  it('header link navigates to home', () => {
    renderApp();
    const headerLink = screen.getByRole('link', { name: /loopa/i });
    expect(headerLink).toHaveAttribute('href', '/');
  });
});
