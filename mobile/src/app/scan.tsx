import { useRef, useState } from 'react';
import { Image, StyleSheet, View } from 'react-native';
import { Button, Card } from 'heroui-native';
import { CameraView, useCameraPermissions } from 'expo-camera';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';

interface CapturedPage {
  uri: string;
  name: string;
}

export default function ScanScreen() {
  const cameraRef = useRef<InstanceType<typeof CameraView>>(null);
  const [permission, requestPermission] = useCameraPermissions();
  const [cameraOpen, setCameraOpen] = useState(false);
  const [capturedPages, setCapturedPages] = useState<CapturedPage[]>([]);
  const [capturing, setCapturing] = useState(false);

  async function openCamera() {
    if (!permission?.granted) {
      const nextPermission = await requestPermission();
      if (!nextPermission.granted) return;
    }

    setCameraOpen(true);
  }

  async function capturePage() {
    if (!cameraRef.current || capturing) return;

    try {
      setCapturing(true);
      const photo = await cameraRef.current.takePictureAsync({ quality: 0.85 });
      if (!photo?.uri) return;

      setCapturedPages((previous) => [
        ...previous,
        { uri: photo.uri, name: `scan-page-${previous.length + 1}.jpg` },
      ]);
      setCameraOpen(false);
    } finally {
      setCapturing(false);
    }
  }

  function removePage(uri: string) {
    setCapturedPages((previous) => previous.filter((page) => page.uri !== uri));
  }

  if (cameraOpen) {
    return (
      <View style={styles.cameraRoot}>
        <CameraView ref={cameraRef} style={styles.camera} facing="back" />
        <View style={styles.cameraActions}>
          <Button variant="secondary" onPress={() => setCameraOpen(false)}>
            <Button.Label>Cancel</Button.Label>
          </Button>
          <Button variant="primary" onPress={() => void capturePage()}>
            <Button.Label>{capturing ? 'Capturing...' : 'Capture'}</Button.Label>
          </Button>
        </View>
      </View>
    );
  }

  return (
    <DocVaultScreen>
      <Card className="rounded-[32px] bg-content1 p-6">
        <View style={styles.cardBody}>
          <ThemedText type="subtitle">Camera scan</ThemedText>
          <ThemedText type="small" themeColor="textSecondary">
            Capture pages, review them, then upload the scan into the same OCR pipeline as web
            uploads.
          </ThemedText>
          {permission && !permission.granted && !permission.canAskAgain && (
            <ThemedText type="small" themeColor="textSecondary">
              Camera permission is blocked. Enable it in system settings to scan documents.
            </ThemedText>
          )}
          <Button variant="primary" onPress={() => void openCamera()}>
            <Button.Label>{capturedPages.length > 0 ? 'Add another page' : 'Open camera'}</Button.Label>
          </Button>
        </View>
      </Card>

      {capturedPages.length > 0 && (
        <Card className="rounded-3xl border border-divider bg-content1 p-4">
          <View style={styles.cardBody}>
            <ThemedText type="smallBold">Captured pages ({capturedPages.length})</ThemedText>
            <View style={styles.previewGrid}>
              {capturedPages.map((page) => (
                <View key={page.uri} style={styles.previewItem}>
                  <Image source={{ uri: page.uri }} style={styles.previewImage} />
                  <Button variant="secondary" size="sm" onPress={() => removePage(page.uri)}>
                    <Button.Label>Remove</Button.Label>
                  </Button>
                </View>
              ))}
            </View>
            <Button variant="primary">
              <Button.Label>Upload scan</Button.Label>
            </Button>
          </View>
        </Card>
      )}
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  cameraRoot: {
    flex: 1,
    backgroundColor: 'black',
  },
  camera: {
    flex: 1,
  },
  cameraActions: {
    position: 'absolute',
    left: Spacing.three,
    right: Spacing.three,
    bottom: Spacing.four,
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: Spacing.three,
  },
  cardBody: {
    gap: Spacing.three,
  },
  previewGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: Spacing.two,
  },
  previewItem: {
    width: '48%',
    gap: Spacing.two,
  },
  previewImage: {
    width: '100%',
    aspectRatio: 3 / 4,
    borderRadius: Spacing.three,
  },
});
