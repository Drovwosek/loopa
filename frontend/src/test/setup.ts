import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock fetch globally
global.fetch = vi.fn();

// Mock clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn().mockResolvedValue(undefined),
    readText: vi.fn().mockResolvedValue(''),
  },
});

// Mock URL.createObjectURL and URL.revokeObjectURL
global.URL.createObjectURL = vi.fn(() => 'blob:mock-url');
global.URL.revokeObjectURL = vi.fn();

// Clean up after each test
afterEach(() => {
  vi.clearAllMocks();
});
