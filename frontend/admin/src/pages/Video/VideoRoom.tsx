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
  LoadingOutlined,
} from '@ant-design/icons';
import TRTC, { Client, Stream, User } from 'trtc-js-sdk';

interface VideoRoomProps {
  caseId: number;
  roomId: string;
  trtcRoomId: number;
  userId: string;
  userSig: string;
  sdkAppId: number;
  userName?: string;
  isHost: boolean;
  onClose: () => void;
}

type TRTCRemoteStream = {
  userId: string;
  userName?: string;
  stream: Stream;
};

const VideoRoom: React.FC<VideoRoomProps> = ({
  caseId,
  roomId,
  trtcRoomId,
  userId: initialUserId,
  userSig: initialUserSig,
  sdkAppId: initialSdkAppId,
  userName = '用户',
  isHost,
  onClose,
}) => {
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
  const [remoteStreams, setRemoteStreams] = useState<TRTCRemoteStream[]>([]);
  const [isJoined, setIsJoined] = useState(false);
  const [isJoining, setIsJoining] = useState(true);
  const [trtcCred, setTrtcCred] = useState<{ sdkAppId: number; userId: string; userSig: string }>({
    sdkAppId: initialSdkAppId,
    userId: initialUserId,
    userSig: initialUserSig,
  });

  const localVideoRef = useRef<HTMLVideoElement>(null);
  const timerRef = useRef<NodeJS.Timeout>();
  const trtcClientRef = useRef<Client | null>(null);
  const localStreamRef = useRef<Stream | null>(null);
  const screenStreamRef = useRef<Stream | null>(null);
  const remoteStreamMapRef = useRef<Map<string, TRTCRemoteStream>>(new Map());

  const fetchTRTCCred = useCallback(async () => {
    try {
      const { videoApi } = await import('../../services/video');
      const resp: any = await videoApi.getTRTCUserSig(caseId, {
        roomId: parseInt(roomId),
        userId: initialUserId || String(Date.now()),
      });
      if (resp && resp.data) {
        setTrtcCred({
          sdkAppId: resp.data.sdkAppId || initialSdkAppId,
          userId: resp.data.userId || initialUserId,
          userSig: resp.data.userSig,
        });
        return resp.data.userSig;
      }
    } catch (err) {
      console.error('Fetch TRTC cred failed:', err);
    }
    return initialUserSig;
  }, [caseId, roomId, initialUserId, initialUserSig, initialSdkAppId]);

  const initTRTC = useCallback(async () => {
    setIsJoining(true);

    let credUserSig = initialUserSig;
    if (!credUserSig || credUserSig.length < 10) {
      credUserSig = await fetchTRTCCred();
    } else {
      setTrtcCred({
        sdkAppId: initialSdkAppId,
        userId: initialUserId,
        userSig: initialUserSig,
      });
    }

    try {
      TRTC.setLogLevel(1);

      const client = TRTC.createClient({
        mode: 'rtc',
        sdkAppId: trtcCred.sdkAppId || initialSdkAppId,
        userId: trtcCred.userId || initialUserId,
        userSig: trtcCred.userSig || credUserSig,
      });

      client.on('stream-added', (event: any) => {
        const remoteStream = event.stream;
        console.log('remote stream added:', remoteStream.getUserId());
        client.subscribe(remoteStream, { audio: true, video: true });
      });

      client.on('stream-subscribed', (event: any) => {
        const remoteStream = event.stream;
        const uid = remoteStream.getUserId();
        console.log('remote stream subscribed:', uid);

        const streamInfo: TRTCRemoteStream = {
          userId: uid,
          stream: remoteStream,
        };
        remoteStreamMapRef.current.set(uid, streamInfo);
        setRemoteStreams(Array.from(remoteStreamMapRef.current.values()));
      });

      client.on('stream-removed', (event: any) => {
        const remoteStream = event.stream;
        const uid = remoteStream.getUserId();
        console.log('remote stream removed:', uid);
        remoteStreamMapRef.current.delete(uid);
        setRemoteStreams(Array.from(remoteStreamMapRef.current.values()));
      });

      client.on('peer-join', (event: any) => {
        console.log('peer join:', event.userId);
        message.info(`用户 ${event.userId} 加入房间`);
      });

      client.on('peer-leave', (event: any) => {
        console.log('peer leave:', event.userId);
        message.warning(`用户 ${event.userId} 离开房间`);
      });

      client.on('error', (error: any) => {
        console.error('trtc client error:', error);
        message.error('TRTC连接错误: ' + error.message);
      });

      await client.join({ roomId: String(trtcRoomId || roomId) });
      console.log('trtc joined room:', roomId);

      const localStream = TRTC.createStream({
        userId: trtcCred.userId || initialUserId,
        audio: true,
        video: true,
      });
      localStream.on('player-state-changed', (event: any) => {
        console.log('local player state:', event);
      });
      await localStream.initialize();
      localStreamRef.current = localStream;

      if (localVideoRef.current) {
        localStream.play(localVideoRef.current);
      }

      await client.publish(localStream);
      console.log('trtc local stream published');

      trtcClientRef.current = client;
      setIsJoined(true);
      setIsJoining(false);
      startCallTimer();
    } catch (err: any) {
      console.error('trtc init failed:', err);
      message.error('加入TRTC房间失败: ' + (err?.message || '未知错误'));
      setIsJoining(false);

      try {
        if (navigator.mediaDevices) {
          const stream = await navigator.mediaDevices.getUserMedia({
            video: true,
            audio: true,
          });
          if (localVideoRef.current) {
            localVideoRef.current.srcObject = stream;
          }
          setIsJoined(true);
          startCallTimer();
        }
      } catch {}
    }
  }, [roomId, trtcRoomId, initialUserId, initialUserSig, initialSdkAppId, fetchTRTCCred, trtcCred]);

  useEffect(() => {
    initTRTC();
    return () => {
      cleanupTRTC();
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [initTRTC]);

  const cleanupTRTC = async () => {
    try {
      if (screenStreamRef.current) {
        screenStreamRef.current.close();
        screenStreamRef.current = null;
      }
      if (localStreamRef.current && trtcClientRef.current) {
        await trtcClientRef.current.unpublish(localStreamRef.current);
        localStreamRef.current.close();
        localStreamRef.current = null;
      }
      if (trtcClientRef.current) {
        await trtcClientRef.current.leave();
        trtcClientRef.current = null;
      }
    } catch (err) {
      console.error('cleanup trtc failed:', err);
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

  const toggleMute = async () => {
    try {
      if (localStreamRef.current) {
        if (isMuted) {
          localStreamRef.current.unmuteAudio();
        } else {
          localStreamRef.current.muteAudio();
        }
        setIsMuted(!isMuted);
      }
    } catch {
      message.error('操作失败');
    }
  };

  const toggleVideo = async () => {
    try {
      if (localStreamRef.current) {
        if (isVideoOff) {
          localStreamRef.current.unmuteVideo();
        } else {
          localStreamRef.current.muteVideo();
        }
        setIsVideoOff(!isVideoOff);
      }
    } catch {
      message.error('操作失败');
    }
  };

  const toggleScreenShare = async () => {
    try {
      if (!trtcClientRef.current) return;

      if (!isScreenSharing) {
        const screenStream = TRTC.createStream({
          userId: trtcCred.userId + '_screen',
          audio: true,
          screen: true,
        });
        await screenStream.initialize();
        await trtcClientRef.current.publish(screenStream);

        screenStreamRef.current = screenStream;

        if (localVideoRef.current) {
          localStreamRef.current?.stop();
          screenStream.play(localVideoRef.current);
        }

        setIsScreenSharing(true);
        message.info('屏幕共享已开启');

        screenStream.on('screen-sharing-stopped', () => {
          toggleScreenShare();
        });
      } else {
        if (screenStreamRef.current) {
          await trtcClientRef.current.unpublish(screenStreamRef.current);
          screenStreamRef.current.close();
          screenStreamRef.current = null;
        }

        if (localVideoRef.current && localStreamRef.current) {
          localStreamRef.current.play(localVideoRef.current);
        }

        setIsScreenSharing(false);
        message.info('屏幕共享已关闭');
      }
    } catch (err: any) {
      console.error('screen share err:', err);
      message.error('屏幕共享操作失败: ' + (err?.message || ''));
    }
  };

  const toggleRecording = async () => {
    try {
      const { videoApi } = await import('../../services/video');
      if (!isRecording) {
        await videoApi.startRecord(caseId, {
          roomId: parseInt(roomId),
          userId: trtcCred.userId || initialUserId,
        });
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
      const { videoApi } = await import('../../services/video');
      if (!isVirtualBg) {
        if (localStreamRef.current && (localStreamRef.current as any).setBackgroundBlur) {
          (localStreamRef.current as any).setBackgroundBlur({
            level: 'high',
          });
        }
        await videoApi.updateVirtualBg(caseId, {
          roomId: parseInt(roomId),
          enabled: true,
        });
        setIsVirtualBg(true);
        message.success('虚拟背景已开启（模糊处理）');
      } else {
        if (localStreamRef.current && (localStreamRef.current as any).setBackgroundBlur) {
          (localStreamRef.current as any).setBackgroundBlur({ level: 'none' });
        }
        await videoApi.updateVirtualBg(caseId, {
          roomId: parseInt(roomId),
          enabled: false,
        });
        setIsVirtualBg(false);
        message.success('虚拟背景已关闭');
      }
    } catch (err: any) {
      message.error('虚拟背景操作失败: ' + (err?.message || ''));
    }
  };

  const toggleBeauty = async () => {
    try {
      const { videoApi } = await import('../../services/video');
      if (!isBeauty) {
        if (localStreamRef.current && (localStreamRef.current as any).setBeauty) {
          (localStreamRef.current as any).setBeauty({
            smooth: 5,
            whiteness: 5,
            thinFace: 3,
            brightEye: 3,
            beautyStyle: 2,
          });
        }
        await videoApi.updateBeauty(caseId, {
          roomId: parseInt(roomId),
          enabled: true,
        });
        setIsBeauty(true);
        message.success('美颜已开启');
      } else {
        if (localStreamRef.current && (localStreamRef.current as any).setBeauty) {
          (localStreamRef.current as any).setBeauty({
            smooth: 0,
            whiteness: 0,
            thinFace: 0,
            brightEye: 0,
          });
        }
        await videoApi.updateBeauty(caseId, {
          roomId: parseInt(roomId),
          enabled: false,
        });
        setIsBeauty(false);
        message.success('美颜已关闭');
      }
    } catch (err: any) {
      message.error('美颜操作失败: ' + (err?.message || ''));
    }
  };

  const generateMinutes = async () => {
    if (!transcriptText.trim()) {
      message.warning('请输入会议转录文本');
      return;
    }
    setIsGeneratingMinutes(true);
    try {
      const { videoApi } = await import('../../services/video');
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
      const { videoApi } = await import('../../services/video');
      const res: any = await videoApi.getMinutes(caseId, roomId);
      setMinutesContent(res.data);
    } catch {
      // No minutes yet
    }
  };

  const fetchQueueList = async () => {
    try {
      const { queueApi } = await import('../../services/video');
      const res: any = await queueApi.getList();
      setQueueList(res.data || []);
    } catch {
      // Ignore
    }
  };

  const handleEndCall = () => {
    Modal.confirm({
      title: '结束视频调解',
      content: '确定要结束本次视频调解吗？结束后将自动停止录制并释放排队资源。',
      okText: '确定结束',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const { videoApi } = await import('../../services/video');
          if (isRecording) {
            await videoApi.stopRecord(caseId, { roomId: parseInt(roomId) });
          }
          await videoApi.endRoom(caseId, roomId);
          await cleanupTRTC();
          message.success('视频调解已结束');
          onClose();
        } catch {
          message.error('结束会议失败');
        }
      },
    });
  };

  const statusMap: Record<number, { color: string; text: string }> = {
    1: { color: 'green', text: '已生成' },
    2: { color: 'blue', text: '已审核' },
    3: { color: 'default', text: '已作废' },
  };

  const statusName = (v: any) => {
    const s = typeof v === 'number' ? statusMap[v] : null;
    return s ? <Tag color={s.color}>{s.text}</Tag> : <Tag>{String(v || '-')}</Tag>;
  };

  const renderRemoteVideos = () => {
    if (remoteStreams.length === 0) {
      return (
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          gap: 8,
        }}>
          {[0, 1].map((i) => (
            <div key={i} style={{
              width: 240,
              height: 135,
              background: 'rgba(15,52,96,0.6)',
              borderRadius: 8,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              position: 'relative',
              flexDirection: 'column',
              color: '#8892b0',
            }}>
              <VideoCameraOutlined style={{ fontSize: 28, marginBottom: 6, opacity: 0.5 }} />
              <div style={{ fontSize: 12 }}>等待参与人加入...</div>
            </div>
          ))}
        </div>
      );
    }

    return (
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
      }}>
        {remoteStreams.map((rs) => (
          <div
            key={rs.userId}
            ref={(el) => {
              if (el && !el.querySelector('video')) {
                try {
                  rs.stream.play(el);
                } catch (e) {
                  console.warn('play remote stream err', e);
                }
              }
            }}
            style={{
              width: 240,
              height: 135,
              background: '#0f3460',
              borderRadius: 8,
              overflow: 'hidden',
              position: 'relative',
            }}
          >
            <Tag
              style={{ position: 'absolute', bottom: 4, left: 4, zIndex: 10 }}
              color="purple"
            >
              {rs.userName || rs.userId}
            </Tag>
          </div>
        ))}
      </div>
    );
  };

  if (isJoining) {
    return (
      <div style={{
        height: '100%',
        background: '#1a1a2e',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        flexDirection: 'column',
        color: '#eaeaea',
      }}>
        <Spin indicator={<LoadingOutlined style={{ fontSize: 48, color: '#533483' }} spin />} size="large" />
        <p style={{ marginTop: 24, fontSize: 16 }}>正在连接视频调解室...</p>
        <p style={{ fontSize: 12, color: '#8892b0', marginTop: 8 }}>
          房间号: {roomId} | 用户: {userName || trtcCred.userId}
        </p>
      </div>
    );
  }

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#1a1a2e' }}>
      <div style={{
        padding: '8px 16px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        background: '#16213e',
        borderBottom: '1px solid #0f3460',
      }}>
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
          {isVirtualBg && <Tag color="purple">虚拟背景</Tag>}
          {isBeauty && <Tag color="magenta">美颜</Tag>}
        </Space>
        <Space>
          <Tooltip title="排队管理">
            <Button
              icon={<TeamOutlined />}
              onClick={() => { fetchQueueList(); setQueueDrawerVisible(true); }}
              style={{ background: 'transparent', color: '#fff', border: '1px solid #0f3460' }}
            />
          </Tooltip>
          <Tooltip title="AI会议纪要">
            <Button
              icon={<FileTextOutlined />}
              onClick={() => { fetchMinutes(); setMinutesDrawerVisible(true); }}
              style={{ background: 'transparent', color: '#fff', border: '1px solid #0f3460' }}
            />
          </Tooltip>
        </Space>
      </div>

      <div style={{ flex: 1, display: 'flex', position: 'relative' }}>
        <div style={{
          flex: 1,
          position: 'relative',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: '#0d1117',
        }}>
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
              objectFit: 'contain',
              background: '#0d1117',
            }}
          />
          <div style={{ position: 'absolute', top: 16, left: 16 }}>
            <Tag color={isScreenSharing ? 'cyan' : 'blue'}>
              {isScreenSharing ? '屏幕共享中' : `${userName || '我'} (${isJoined ? '已连接' : '未连接'})`}
            </Tag>
          </div>
          {remoteStreams.length > 0 && (
            <div style={{ position: 'absolute', top: 16, right: 272 }}>
              <Badge count={remoteStreams.length} showZero>
                <Tag color="purple">远端用户</Tag>
              </Badge>
            </div>
          )}
        </div>

        <div style={{
          position: 'absolute',
          top: 16,
          right: 16,
          display: 'flex',
          flexDirection: 'column',
          gap: 8,
        }}>
          {renderRemoteVideos()}
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
        title="AI会议纪要 (DeepSeek)"
        placement="right"
        width={500}
        open={minutesDrawerVisible}
        onClose={() => setMinutesDrawerVisible(false)}
      >
        {!minutesContent ? (
          <div>
            <div style={{ marginBottom: 16 }}>
              <Input.TextArea
                rows={10}
                placeholder="请粘贴或输入会议转录文本。DeepSeek将自动提取: 争议焦点、调解过程、达成协议、风险提示、下一步行动..."
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
              <Descriptions.Item label="会议标题">{minutesContent.meeting_title || '-'}</Descriptions.Item>
              <Descriptions.Item label="概要">{minutesContent.summary || '-'}</Descriptions.Item>
              <Descriptions.Item label="争议焦点">
                {(() => {
                  try {
                    const arr = typeof minutesContent.dispute_focus === 'string' ? JSON.parse(minutesContent.dispute_focus) : [];
                    return Array.isArray(arr) && arr.length > 0
                      ? arr.map((f: string, i: number) => <Tag key={i} color="orange">{f}</Tag>)
                      : String(minutesContent.dispute_focus || '-');
                  } catch { return String(minutesContent.dispute_focus || '-'); }
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="达成协议">{minutesContent.agreement || '-'}</Descriptions.Item>
              <Descriptions.Item label="调解过程" span={2}>
                <div style={{ maxHeight: 180, overflow: 'auto' }}>
                  {minutesContent.mediation_process || '-'}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="风险提示">
                {(() => {
                  try {
                    const arr = typeof minutesContent.risk_points === 'string' ? JSON.parse(minutesContent.risk_points) : [];
                    return Array.isArray(arr) && arr.length > 0
                      ? arr.map((r: string, i: number) => <Tag key={i} color="red">{r}</Tag>)
                      : String(minutesContent.risk_points || '-');
                  } catch { return String(minutesContent.risk_points || '-'); }
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="下一步">
                {(() => {
                  try {
                    const arr = typeof minutesContent.next_steps === 'string' ? JSON.parse(minutesContent.next_steps) : [];
                    return Array.isArray(arr) && arr.length > 0
                      ? arr.map((s: string, i: number) => <Tag key={i} color="blue">{s}</Tag>)
                      : String(minutesContent.next_steps || '-');
                  } catch { return String(minutesContent.next_steps || '-'); }
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="情绪状态">{minutesContent.emotional_state || '-'}</Descriptions.Item>
              <Descriptions.Item label="调解员建议">{minutesContent.mediator_advice || '-'}</Descriptions.Item>
              <Descriptions.Item label="状态">{statusName(minutesContent.status)}</Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Drawer>

      <Drawer
        title="排队管理"
        placement="right"
        width={420}
        open={queueDrawerVisible}
        onClose={() => setQueueDrawerVisible(false)}
      >
        <List
          dataSource={queueList}
          locale={{ emptyText: '当前无排队人员' }}
          renderItem={(item: any, index: number) => (
            <List.Item>
              <List.Item.Meta
                avatar={
                  <Badge
                    count={index + 1}
                    style={{
                      backgroundColor:
                        item.priority === 1 ? '#ff4d4f' :
                        item.priority === 2 ? '#faad14' : '#1890ff',
                    }}
                  />
                }
                title={
                  <Space>
                    <span>{item.partyName}</span>
                    {item.priority === 1 && <Tag color="red">特急</Tag>}
                    {item.priority === 2 && <Tag color="gold">紧急</Tag>}
                  </Space>
                }
                description={
                  <div>
                    <div>调解员: {item.mediatorName || '-'}</div>
                    <div>案件: {item.caseNo || '-'}</div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </Drawer>

      <style>{`
        @keyframes blink {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
        video { outline: none; }
      `}</style>
    </div>
  );
};

export default VideoRoom;
