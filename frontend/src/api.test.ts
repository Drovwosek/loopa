import { describe, it, expect, vi, beforeEach } from 'vitest';
import { uploadFile, fetchTask, fetchHistory, deleteTask, downloadExport } from './api';

describe('API functions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('uploadFile', () => {
    it('should upload file and return taskId', async () => {
      const mockResponse = { taskId: 'task-123' };
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      } as Response);

      const file = new File(['test content'], 'test.mp3', { type: 'audio/mpeg' });
      const result = await uploadFile(file);

      expect(result).toBe('task-123');
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/uploads'),
        expect.objectContaining({
          method: 'POST',
          credentials: 'include',
        })
      );
    });

    it('should throw error on upload failure', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: false,
        json: () => Promise.resolve({ error: 'File too large' }),
      } as Response);

      const file = new File(['test content'], 'test.mp3', { type: 'audio/mpeg' });

      await expect(uploadFile(file)).rejects.toThrow('File too large');
    });

    it('should throw default error when no error message', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: false,
        json: () => Promise.resolve({}),
      } as Response);

      const file = new File(['test content'], 'test.mp3', { type: 'audio/mpeg' });

      await expect(uploadFile(file)).rejects.toThrow('Upload failed');
    });
  });

  describe('fetchTask', () => {
    it('should fetch and return task', async () => {
      const mockTask = {
        id: 'task-123',
        status: 'готово',
        originalName: 'test.mp3',
        transcriptText: 'Hello world',
        createdAt: '2024-01-01T00:00:00Z',
      };

      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockTask),
      } as Response);

      const result = await fetchTask('task-123');

      expect(result).toEqual(mockTask);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/tasks/task-123'),
        expect.objectContaining({
          credentials: 'include',
        })
      );
    });

    it('should throw error when task not found', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: false,
      } as Response);

      await expect(fetchTask('nonexistent')).rejects.toThrow('Failed to load task');
    });
  });

  describe('fetchHistory', () => {
    it('should fetch and return history items', async () => {
      const mockHistory = [
        { id: 'task-1', originalName: 'file1.mp3', status: 'готово', uploadedAt: '2024-01-01T00:00:00Z' },
        { id: 'task-2', originalName: 'file2.wav', status: 'ожидает', uploadedAt: '2024-01-02T00:00:00Z' },
      ];

      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockHistory),
      } as Response);

      const result = await fetchHistory();

      expect(result).toEqual(mockHistory);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/history'),
        expect.objectContaining({
          credentials: 'include',
        })
      );
    });

    it('should return empty array when no history', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve([]),
      } as Response);

      const result = await fetchHistory();
      expect(result).toEqual([]);
    });

    it('should throw error on failure', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: false,
      } as Response);

      await expect(fetchHistory()).rejects.toThrow('Failed to load history');
    });
  });

  describe('deleteTask', () => {
    it('should delete task successfully', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
      } as Response);

      await expect(deleteTask('task-123')).resolves.toBeUndefined();
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/tasks/task-123'),
        expect.objectContaining({
          method: 'DELETE',
          credentials: 'include',
        })
      );
    });

    it('should throw error on failure', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: false,
      } as Response);

      await expect(deleteTask('task-123')).rejects.toThrow('Failed to delete task');
    });
  });

  describe('downloadExport', () => {
    it('should download TXT export', async () => {
      const mockBlob = new Blob(['transcript text'], { type: 'text/plain' });
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        blob: () => Promise.resolve(mockBlob),
        headers: new Headers({
          'content-disposition': 'attachment; filename="test.txt"',
        }),
      } as Response);

      // Mock click and link creation
      const mockClick = vi.fn();
      vi.spyOn(document, 'createElement').mockReturnValue({
        href: '',
        download: '',
        click: mockClick,
      } as unknown as HTMLAnchorElement);

      await downloadExport('task-123', 'txt');

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/tasks/task-123/export?format=txt'),
        expect.objectContaining({
          credentials: 'include',
        })
      );
      expect(mockClick).toHaveBeenCalled();
    });

    it('should download DOCX export', async () => {
      const mockBlob = new Blob(['docx content'], {
        type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      });
      vi.mocked(global.fetch).mockResolvedValue({
        ok: true,
        blob: () => Promise.resolve(mockBlob),
        headers: new Headers({
          'content-disposition': 'attachment; filename="test.docx"',
        }),
      } as Response);

      const mockClick = vi.fn();
      vi.spyOn(document, 'createElement').mockReturnValue({
        href: '',
        download: '',
        click: mockClick,
      } as unknown as HTMLAnchorElement);

      await downloadExport('task-123', 'docx');

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/tasks/task-123/export?format=docx'),
        expect.objectContaining({
          credentials: 'include',
        })
      );
    });

    it('should throw error on export failure', async () => {
      vi.mocked(global.fetch).mockResolvedValue({
        ok: false,
      } as Response);

      await expect(downloadExport('task-123', 'txt')).rejects.toThrow('Export failed');
    });
  });
});
