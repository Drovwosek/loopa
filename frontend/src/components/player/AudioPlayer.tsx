import { useEffect, useRef, useState, useCallback } from "react";
import { Card, Button, Space, Typography } from "antd";
import {
  PlayCircleOutlined,
  PauseOutlined,
  StepBackwardOutlined,
} from "@ant-design/icons";
import WaveSurfer from "wavesurfer.js";

const { Text } = Typography;

type AudioPlayerProps = {
  audioUrl: string;
  onTimeUpdate?: (timeMs: number) => void;
};

function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export default function AudioPlayer({ audioUrl, onTimeUpdate }: AudioPlayerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const wavesurferRef = useRef<WaveSurfer | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);

  useEffect(() => {
    if (!containerRef.current) return;

    const ws = WaveSurfer.create({
      container: containerRef.current,
      waveColor: "#bfbfbf",
      progressColor: "#1677ff",
      cursorColor: "#1677ff",
      barWidth: 2,
      barRadius: 3,
      cursorWidth: 1,
      height: 80,
      barGap: 2,
    });

    ws.load(audioUrl);

    ws.on("ready", () => {
      setDuration(ws.getDuration());
    });

    ws.on("audioprocess", () => {
      const time = ws.getCurrentTime();
      setCurrentTime(time);
      onTimeUpdate?.(time * 1000);
    });

    ws.on("seeking", () => {
      const time = ws.getCurrentTime();
      setCurrentTime(time);
      onTimeUpdate?.(time * 1000);
    });

    ws.on("play", () => setIsPlaying(true));
    ws.on("pause", () => setIsPlaying(false));

    wavesurferRef.current = ws;

    return () => {
      ws.destroy();
    };
  }, [audioUrl]);

  const togglePlay = useCallback(() => {
    wavesurferRef.current?.playPause();
  }, []);

  const seekTo = useCallback((timeMs: number) => {
    const ws = wavesurferRef.current;
    if (ws && duration > 0) {
      ws.seekTo(timeMs / 1000 / duration);
      ws.play();
    }
  }, [duration]);

  // Expose seekTo to parent via ref pattern
  useEffect(() => {
    (window as any).__loopaAudioSeek = seekTo;
    return () => {
      delete (window as any).__loopaAudioSeek;
    };
  }, [seekTo]);

  return (
    <Card size="small" style={{ marginBottom: 16 }}>
      <div ref={containerRef} style={{ marginBottom: 12 }} />
      <Space>
        <Button
          icon={isPlaying ? <PauseOutlined /> : <PlayCircleOutlined />}
          onClick={togglePlay}
        >
          {isPlaying ? "Пауза" : "Воспроизвести"}
        </Button>
        <Button
          icon={<StepBackwardOutlined />}
          onClick={() => {
            wavesurferRef.current?.seekTo(0);
            setCurrentTime(0);
          }}
        >
          В начало
        </Button>
        <Text type="secondary">
          {formatTime(currentTime)} / {formatTime(duration)}
        </Text>
      </Space>
    </Card>
  );
}
