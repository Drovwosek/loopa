import React from "react";
import { Layout, Menu, Typography } from "antd";
import { SoundOutlined, FolderOutlined, HistoryOutlined } from "@ant-design/icons";
import { Link, useLocation } from "react-router-dom";

const { Header, Content, Sider } = Layout;
const { Title } = Typography;

type AppLayoutProps = {
  children: React.ReactNode;
};

export default function AppLayout({ children }: AppLayoutProps) {
  const location = useLocation();

  const menuItems = [
    {
      key: "/",
      icon: <SoundOutlined />,
      label: <Link to="/">Транскрибация</Link>,
    },
    {
      key: "/projects",
      icon: <FolderOutlined />,
      label: <Link to="/projects">Проекты</Link>,
    },
  ];

  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Sider
        breakpoint="lg"
        collapsedWidth={0}
        style={{ background: "#fff" }}
      >
        <div style={{ padding: "16px 24px" }}>
          <Title level={4} style={{ margin: 0 }}>
            Loopa
          </Title>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
        />
      </Sider>
      <Layout>
        <Content style={{ margin: 24 }}>{children}</Content>
      </Layout>
    </Layout>
  );
}
