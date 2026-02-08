import { useState } from "react";
import { Upload, message, Card, Button } from "antd";
import { InboxOutlined } from "@ant-design/icons";

const { Dragger } = Upload;

type FileUploadProps = {
  onUploadStart: (file: File) => void;
  uploading: boolean;
};

export default function FileUpload({ onUploadStart, uploading }: FileUploadProps) {
  const [file, setFile] = useState<File | null>(null);

  return (
    <Card title="Загрузка медиафайла">
      <Dragger
        name="file"
        accept=".mp3,.wav,.mp4,.mov,audio/*,video/*"
        multiple={false}
        showUploadList={false}
        beforeUpload={(f) => {
          const isLt1G = f.size / 1024 / 1024 / 1024 < 1;
          if (!isLt1G) {
            message.error("Файл должен быть меньше 1 ГБ");
            return false;
          }
          setFile(f);
          return false; // Не загружаем автоматически
        }}
      >
        <p className="ant-upload-drag-icon">
          <InboxOutlined />
        </p>
        <p className="ant-upload-text">
          Нажмите или перетащите файл для загрузки
        </p>
        <p className="ant-upload-hint">
          MP3, WAV, MP4, MOV до 1 ГБ
        </p>
      </Dragger>

      {file && (
        <div style={{ marginTop: 16, textAlign: "center" }}>
          <p>Выбран: <strong>{file.name}</strong></p>
          <Button
            type="primary"
            loading={uploading}
            onClick={() => onUploadStart(file)}
          >
            Начать транскрибацию
          </Button>
        </div>
      )}
    </Card>
  );
}
