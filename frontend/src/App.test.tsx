import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';
import { configureStore } from '@reduxjs/toolkit';
import App from './App';
import historyReducer from './store/historySlice';
import taskReducer from './store/taskSlice';
import projectReducer from './store/projectSlice';
import * as api from './api';

vi.mock('./api');

const createTestStore = (preloadedState = {}) =>
  configureStore({
    reducer: {
      history: historyReducer,
      task: taskReducer,
      projects: projectReducer,
    },
    preloadedState,
  });

const renderApp = (initialEntries = ['/'], preloadedState = {}) => {
  vi.mocked(api.fetchHistory).mockResolvedValue([]);
  vi.mocked(api.fetchProjects).mockResolvedValue([]);
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
  it('renders the Loopa title in sidebar', () => {
    renderApp();
    expect(screen.getByText('Loopa')).toBeInTheDocument();
  });

  it('renders menu items', () => {
    renderApp();
    expect(screen.getByText('Транскрибация')).toBeInTheDocument();
    expect(screen.getByText('Проекты')).toBeInTheDocument();
  });

  it('renders HomePage at root route', () => {
    renderApp(['/']);
    expect(screen.getByText('Загрузка медиафайла')).toBeInTheDocument();
  });

  it('renders ProjectPage at /projects route', () => {
    vi.mocked(api.fetchProjects).mockResolvedValue([]);
    renderApp(['/projects']);
    // Title "Проекты" in PageHeader (h4)
    expect(screen.getByRole('heading', { name: 'Проекты' })).toBeInTheDocument();
  });
});
