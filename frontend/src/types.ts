export type Segment = {
  id: string;
  speakerId?: string;
  speakerName?: string;
  startTime: number;
  endTime: number;
  text: string;
  hasFillers: boolean;
  isCorrected: boolean;
};

export type Project = {
  id: string;
  name: string;
  description?: string;
  status: string;
  createdAt: string;
  fileCount: number;
};
