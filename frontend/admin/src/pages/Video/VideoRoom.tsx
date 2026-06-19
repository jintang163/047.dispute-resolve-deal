import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  Card,
  Button,
  Space,
  Modal,
  message,
  Tag,
  Spin,
  Tooltip,
  Drawer,
  Descriptions,
  List,
  Input,
  Select,
  Form,
  Row,
  Col,
  Statistic,
  Badge,
  Timeline,
} from 'antd';
import {
  VideoCameraOutlined,
  PhoneOutlined,
  CameraOutlined,
  ShareAltOutlined,
  StopOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  BorderOutlined,
  SmileOutlined,
  FileTextOutlined,
  TeamOutlined,
  QueueCtrlOutlined,
  CloudOutlined,
  SoundOutlined,
} from '@ant-design/icons';

interface VideoRoomProps {
  caseId: number;
  roomId: string;
  trtcRoomId: number;
  userId: string;
  userSig: string;
  sdkAppId: number;
  isHost: boolean;
  onClose: () => void;
}

const VideoRoom: React.FC<VideoRoomProps> = ({
  caseId,
  roomId,
  trtcRoomId,
  userId,
  userSig,
  sdkAppId,
  isHost,
  onClose,
}) => {
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStreams, setRemoteStreams] = useState<Map<string, MediaStream>>(new Map());
  const [isMuted, setIsMuted] = useState(false);
  const [isVideoOff, setIsVideoOff] = useState(false);
  const [isScreenSharing, setIsScreenSharing] = useState(false);
  const [isRecording, setIsRecording] = useState(false);
  const [isVirtualBg, setIsVirtualBg] = useState(false);
  const [isBeauty, setIsBeauty] = useState(false);
  const [minutesDrawerVisible, setMinutesDrawerVisible] = useState(false);
  const [queueDrawerVisible, setQueueDrawerVisible] = useState(false);
  const [minutesContent, setMinutesContent] = useState<any>(null);
  const [queueList, setQueueList] = useState<any[]>([]);
  const [transcriptText, setTranscriptText] = useState('');
  const [isGeneratingMinutes, setIsGeneratingMinutes] = useState(false);
  const [callDuration, setCallDuration] = useState(0);

  const localVideoRef = useRef<HTMLVideoElement>(null);
  const remoteVideoRefs = useRef<Map<string, HTMLVideoElement>>(new Map());
  const timerRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    startLocalStream();
    startCallTimer();
    return () => {
      stopLocalStream();
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  const startLocalStream = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: true,
        audio: true,
      });
      setLocalStream(stream);
      if (localVideoRef.current) {
        localVideoRef.current.srcObject = stream;
      }
      message.success('摄像头和麦克风已启用');
    } catch (err) {
      message.error('无法访问摄像头或麦克风');
      console.error('getUserMedia error:', err);
    }
  };

  const stopLocalStream = () => {
    if (localStream) {
      localStream.getTracks().forEach((track) => track.stop());
    }
  };

  const startCallTimer = () => {
    timerRef.current = setInterval(() => {
      setCallDuration((prev) => prev + 1);
    }, 1000);
  };

  const formatDuration = (seconds: number) => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
    return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  };

  const toggleMute = () => {
    if (localStream) {
      localStream.getAudioTracks().forEach((track) => {
        track.enabled = !track.enabled;
      });
      setIsMuted(!isMuted);
    }
  };

  const toggleVideo = () => {
    if (localStream) {
      localStream.getVideoTracks().forEach((track) => {
        track.enabled = !track.enabled;
      });
      setIsVideoOff(!isVideoOff);
    }
  };

  const toggleScreenShare = async () => {
    if (!isScreenSharing) {
      try {
        const screenStream = await navigator.mediaDevices.getDisplayMedia({
          video: true,
          audio: true,
        });
        if (localVideoRef.current) {
          localVideoRef.current.srcObject = screenStream;
        }
        setIsScreenSharing(true);
        message.info('屏幕共享已开启');

        screenStream.getVideoTracks()[0].onended = () => {
          if (localVideoRef.current && localStream) {
            localVideoRef.current.srcObject = localStream;
          }
          setIsScreenSharing(false);
          message.info('屏幕共享已关闭');
        };
      } catch {
        message.error('无法启动屏幕共享');
      }
    } else {
      if (localVideoRef.current && localStream) {
        localVideoRef.current.srcObject = localStream;
      }
      setIsScreenSharing(false);
      message.info('屏幕共享已关闭');
    }
  };

  const toggleRecording = async () => {
    try {
      const { videoApi } = await import('../services/video');
      if (!isRecording) {
        await videoApi.startRecord(caseId, { roomId: parseInt(roomId), userId });
        setIsRecording(true);
        message.success('云端录制已启动');
      } else {
        await videoApi.stopRecord(caseId, { roomId: parseInt(roomId) });
        setIsRecording(false);
        message.success('云端录制已停止');
      }
    } catch {
      message.error('录制操作失败');
    }
  };

  const toggleVirtualBg = async () => {
    try {
      const { videoApi } = await import('../services/video');
      await videoApi.updateVirtualBg(caseId, { roomId: parseInt(roomId), enabled: !isVirtualBg });
      setIsVirtualBg(!isVirtualBg);
      message.success(isVirtualBg ? '虚拟背景已关闭' : '虚拟背景已开启');
    } catch {
      message.error('虚拟背景操作失败');
    }
  };

  const toggleBeauty = async () => {
    try {
      const { videoApi } = await import('../services/video');
      await videoApi.updateBeauty(caseId, { roomId: parseInt(roomId), enabled: !isBeauty });
      setIsBeauty(!isBeauty);
      message.success(isBeauty ? '美颜已关闭' : '美颜已开启');
    } catch {
      message.error('美颜操作失败');
    }
  };

  const generateMinutes = async () => {
    if (!transcriptText.trim()) {
      message.warning('请输入会议转录文本');
      return;
    }
    setIsGeneratingMinutes(true);
    try {
      const { videoApi } = await import('../services/video');
      const res: any = await videoApi.generateMinutes(caseId, {
        roomId: parseInt(roomId),
        caseId,
        transcript: transcriptText,
        durationMinutes: Math.ceil(callDuration / 60),
      });
      message.success('会议纪要生成成功');
      fetchMinutes();
    } catch {
      message.error('生成会议纪要失败');
    } finally {
      setIsGeneratingMinutes(false);
    }
  };

  const fetchMinutes = async () => {
    try {
      const { videoApi } = await import('../services/video');
      const res: any = await videoApi.getMinutes(caseId, roomId);
      setMinutesContent(res.data);
    } catch {
      // No minutes yet
    }
  };

  const fetchQueueList = async () => {
    try {
      const { queueApi } = await import('../services/video');
      const res: any = await queueApi.getList();
      setQueueList(res.data || []);
    } catch {
      // Ignore
    }
  };

  const handleEndCall = () => {
    Modal.confirm({
      title: '结束视频调解',
      content: '确定要结束本次视频调解吗？结束后将自动停止录制。',
      okText: '确定结束',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const { videoApi } = await import('../services/video');
          if (isRecording) {
            await videoApi.stopRecord(caseId, { roomId: parseInt(roomId) });
          }
          await videoApi.endRoom(caseId, roomId);
          stopLocalStream();
          message.success('视频调解已结束');
          onClose();
        } catch {
          message.error('结束会议失败');
        }
      },
    });
  };

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#1a1a2e' }}>
      <div style={{ padding: '8px 16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', background: '#16213e', borderBottom: '1px solid #0f3460' }}>
        <Space>
          <Tag color="green" style={{ fontSize: 14 }}>
            <VideoCameraOutlined /> 视频调解中
          </Tag>
          <Tag color="blue" style={{ fontSize: 14 }}>
            {formatDuration(callDuration)}
          </Tag>
          {isRecording && (
            <Tag color="red" style={{ fontSize: 14, animation: 'blink 1s infinite' }}>
              <CloudOutlined /> 录制中
            </Tag>
          )}
        </Space>
        <Space>
          <Tooltip title="排队管理">
            <Button icon={<TeamOutlined />} onClick={() => { fetchQueueList(); setQueueDrawerVisible(true); }} style={{ background: 'transparent', color: '#fff', border: '1px solid #0f3460' }} />
          </Tooltip>
          <Tooltip title="会议纪要">
            <Button icon={<FileTextOutlined />} onClick={() => { fetchMinutes(); setMinutesDrawerVisible(true); }} style={{ background: 'transparent', color: '#fff', border: '1px solid #0f3460' }} />
          </Tooltip>
        </Space>
      </div>

      <div style={{ flex: 1, display: 'flex', position: 'relative' }}>
        <div style={{ flex: 1, position: 'relative', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <video
            ref={localVideoRef}
            autoPlay
            muted
            playsInline
            style={{
              width: isScreenSharing ? '100%' : 'auto',
              maxWidth: '100%',
              maxHeight: '100%',
              borderRadius: 8,
              transform: isScreenSharing ? 'none' : 'scaleX(-1)',
            }}
          />
          <div style={{ position: 'absolute', top: 16, left: 16 }}>
            <Tag color="blue">{isScreenSharing ? '屏幕共享中' : '我'}</Tag>
          </div>
        </div>

        <div style={{
          position: 'absolute',
          top: 16,
          right: 16,
          width: 240,
          display: 'flex',
          flexDirection: 'column',
          gap: 8,
        }}>
          {[1, 2].map((idx) => (
            <div key={idx} style={{
              width: 240,
              height: 135,
              background: '#0f3460',
              borderRadius: 8,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              position: 'relative',
            }}>
              <VideoCameraOutlined style={{ fontSize: 32, color: '#533483' }} />
              <Tag style={{ position: 'absolute', bottom: 4, left: 4 }} color="purple">
                参与人 {idx}
              </Tag>
            </div>
          ))}
        </div>
      </div>

      <div style={{
        padding: '12px 16px',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        gap: 8,
        background: '#16213e',
        borderTop: '1px solid #0f3460',
      }}>
        <Tooltip title={isMuted ? '取消静音' : '静音'}>
          <Button
            type={isMuted ? 'primary' : 'default'}
            danger={isMuted}
            shape="circle"
            size="large"
            icon={<SoundOutlined />}
            onClick={toggleMute}
          />
        </Tooltip>

        <Tooltip title={isVideoOff ? '开启摄像头' : '关闭摄像头'}>
          <Button
            type={isVideoOff ? 'primary' : 'default'}
            danger={isVideoOff}
            shape="circle"
            size="large"
            icon={<CameraOutlined />}
            onClick={toggleVideo}
          />
        </Tooltip>

        <Tooltip title={isScreenSharing ? '停止共享' : '屏幕共享'}>
          <Button
            type={isScreenSharing ? 'primary' : 'default'}
            shape="circle"
            size="large"
            icon={<ShareAltOutlined />}
            onClick={toggleScreenShare}
          />
        </Tooltip>

        {isHost && (
          <>
            <Tooltip title={isRecording ? '停止录制' : '开始录制'}>
              <Button
                type={isRecording ? 'primary' : 'default'}
                danger={isRecording}
                shape="circle"
                size="large"
                icon={<CloudOutlined />}
                onClick={toggleRecording}
              />
            </Tooltip>

            <Tooltip title={isVirtualBg ? '关闭虚拟背景' : '虚拟背景'}>
              <Button
                type={isVirtualBg ? 'primary' : 'default'}
                shape="circle"
                size="large"
                icon={<BorderOutlined />}
                onClick={toggleVirtualBg}
              />
            </Tooltip>

            <Tooltip title={isBeauty ? '关闭美颜' : '美颜'}>
              <Button
                type={isBeauty ? 'primary' : 'default'}
                shape="circle"
                size="large"
                icon={<SmileOutlined />}
                onClick={toggleBeauty}
              />
            </Tooltip>
          </>
        )}

        <Tooltip title="AI会议纪要">
          <Button
            shape="circle"
            size="large"
            icon={<FileTextOutlined />}
            onClick={() => { fetchMinutes(); setMinutesDrawerVisible(true); }}
          />
        </Tooltip>

        <Tooltip title="结束通话">
          <Button
            type="primary"
            danger
            shape="circle"
            size="large"
            icon={<PhoneOutlined style={{ transform: 'rotate(135deg)' }} />}
            onClick={handleEndCall}
          />
        </Tooltip>
      </div>

      <Drawer
        title="AI会议纪要"
        placement="right"
        width={500}
        open={minutesDrawerVisible}
        onClose={() => setMinutesDrawerVisible(false)}
      >
        {!minutesContent ? (
          <div>
            <div style={{ marginBottom: 16 }}>
              <Input.TextArea
                rows={8}
                placeholder="请输入会议转录文本，DeepSeek将自动生成结构化会议纪要..."
                value={transcriptText}
                onChange={(e) => setTranscriptText(e.target.value)}
              />
            </div>
            <Button
              type="primary"
              icon={<FileTextOutlined />}
              loading={isGeneratingMinutes}
              onClick={generateMinutes}
              block
            >
              DeepSeek生成会议纪要
            </Button>
          </div>
        ) : (
          <div>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="会议标题">{minutesContent.meeting_title}</Descriptions.Item>
              <Descriptions.Item label="会议概要">{minutesContent.summary}</Descriptions.Item>
              <Descriptions.Item label="争议焦点">
                {Array.isArray(minutesContent.dispute_focus)
                  ? minutesContent.dispute_focus.map((f: string, i: number) => <Tag key={i} color="orange">{f}</Tag>)
                  : minutesContent.dispute_focus}
              </Descriptions.Item>
              <Descriptions.Item label="达成协议">{minutesContent.agreement}</Descriptions.Item>
              <Descriptions.Item label="调解过程">{minutesContent.mediation_process}</Descriptions.Item>
              <Descriptions.Item label="风险提示">
                {Array.isArray(minutesContent.risk_points)
                  ? minutesContent.risk_points.map((r: string, i: number) => <Tag key={i} color="red">{r}</Tag>)
                  : minutesContent.risk_points}
              </Descriptions.Item>
              <Descriptions.Item label="下一步行动">
                {Array.isArray(minutesContent.next_steps)
                  ? minutesContent.next_steps.map((s: string, i: number) => <Tag key={i} color="blue">{s}</Tag>)
                  : minutesContent.next_steps}
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Drawer>

      <Drawer
        title="排队管理"
        placement="right"
        width={400}
        open={queueDrawerVisible}
        onClose={() => setQueueDrawerVisible(false)}
      >
        <List
          dataSource={queueList}
          renderItem={(item: any, index: number) => (
            <List.Item>
              <List.Item.Meta
                avatar={<Badge count={index + 1} style={{ backgroundColor: item.priority === 1 ? '#ff4d4f' : item.priority === 2 ? '#faad14' : '#1890ff' }} />}
                title={item.partyName}
                description={`调解员: ${item.mediatorName} | 案件: ${item.caseNo}`}
              />
            </List.Item>
          )}
          locale={{ emptyText: '当前无排队人员' }}
        />
      </Drawer>

      <style>{`
        @keyframes blink {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
      `}</style>
    </div>
  );
};

export default VideoRoom;
