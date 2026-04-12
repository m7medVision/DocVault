import { useRef, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Image,
  ScrollView,
  TextInput,
  Modal,
  KeyboardAvoidingView,
  Platform,
  Animated,
} from 'react-native';
import { useRouter } from 'expo-router';
import { CameraView } from 'expo-camera';
import { useAuth } from '../../src/contexts/AuthContext';
import { tokens, colors } from '../../src/theme/tokens';
import { useCameraCapture, DOCUMENT_TYPES } from '../../src/features/documents/useCameraCapture';

export default function CameraCaptureScreen() {
  const router = useRouter();
  const { accessToken } = useAuth();
  const captureAnim = useRef(new Animated.Value(1)).current;

  const {
    cameraPermission,
    cameraRef,
    facing,
    flash,
    mode,
    capturedImage,
    uploadProgress,
    uploadError,
    title,
    docType,
    toggleFacing,
    toggleFlash,
    capturePhoto,
    retakePhoto,
    uploadDocument,
    setTitle,
    setDocType,
    requestCameraPermission,
  } = useCameraCapture(accessToken);

  useEffect(() => {
    if (mode === 'capturing') {
      Animated.sequence([
        Animated.timing(captureAnim, {
          toValue: 0.8,
          duration: 100,
          useNativeDriver: true,
        }),
        Animated.timing(captureAnim, {
          toValue: 1,
          duration: 100,
          useNativeDriver: true,
        }),
      ]).start();
    }
  }, [mode]);

  const goToDocuments = () => {
    router.replace('/(app)/documents');
  };

  if (cameraPermission?.status === 'denied') {
    return (
      <View style={styles.permissionContainer}>
        <Text style={styles.permissionIcon}>📷</Text>
        <Text style={styles.permissionTitle}>Camera Access Required</Text>
        <Text style={styles.permissionText}>
          DocVault needs camera access to capture document photos.
          Please enable camera access in your device settings.
        </Text>
        <TouchableOpacity style={styles.permissionButton} onPress={requestCameraPermission}>
          <Text style={styles.permissionButtonText}>Grant Permission</Text>
        </TouchableOpacity>
        <TouchableOpacity style={[styles.permissionButton, styles.secondaryButton]} onPress={() => router.back()}>
          <Text style={styles.secondaryButtonText}>Go Back</Text>
        </TouchableOpacity>
      </View>
    );
  }

  if (!cameraPermission) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={tokens.accentColor} />
        <Text style={styles.loadingText}>Initializing camera...</Text>
      </View>
    );
  }

  if (mode === 'preview' && capturedImage) {
    return (
      <View style={styles.container}>
        <Image source={{ uri: capturedImage }} style={styles.previewImage} resizeMode="contain" />

        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          style={styles.formContainer}
        >
          <ScrollView contentContainerStyle={styles.formContent}>
            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>Document Title</Text>
              <TextInput
                style={styles.textInput}
                value={title}
                onChangeText={setTitle}
                placeholder="Enter document title"
                placeholderTextColor={colors.gray[500]}
                autoCapitalize="words"
              />
            </View>

            <View style={styles.inputGroup}>
              <Text style={styles.inputLabel}>Document Type</Text>
              <TouchableOpacity style={styles.typeSelector} onPress={() => {}}>
                <Text style={styles.typeSelectorText}>
                  {DOCUMENT_TYPES.find((t) => t.value === docType)?.icon}{' '}
                  {DOCUMENT_TYPES.find((t) => t.value === docType)?.label}
                </Text>
                <Text style={styles.typeSelectorArrow}>▼</Text>
              </TouchableOpacity>
            </View>

            {uploadError && (
              <View style={styles.errorContainer}>
                <Text style={styles.errorText}>{uploadError}</Text>
              </View>
            )}
          </ScrollView>

          <View style={styles.previewActions}>
            <TouchableOpacity style={styles.retakeButton} onPress={retakePhoto}>
              <Text style={styles.retakeButtonText}>Retake</Text>
            </TouchableOpacity>
            <TouchableOpacity style={styles.uploadButton} onPress={uploadDocument}>
              <Text style={styles.uploadButtonText}>Upload Document</Text>
            </TouchableOpacity>
          </View>
        </KeyboardAvoidingView>

        <Modal visible={false} transparent animationType="slide">
          <TouchableOpacity style={styles.modalOverlay} activeOpacity={1}>
            <View style={styles.modalContent}>
              <View style={styles.modalHeader}>
                <Text style={styles.modalTitle}>Select Document Type</Text>
                <TouchableOpacity>
                  <Text style={styles.modalClose}>✕</Text>
                </TouchableOpacity>
              </View>
              {DOCUMENT_TYPES.map((type) => (
                <TouchableOpacity
                  key={type.value}
                  style={[
                    styles.typeOption,
                    docType === type.value && styles.typeOptionActive,
                  ]}
                  onPress={() => setDocType(type.value)}
                >
                  <Text style={styles.typeOptionIcon}>{type.icon}</Text>
                  <Text
                    style={[
                      styles.typeOptionText,
                      docType === type.value && styles.typeOptionTextActive,
                    ]}
                  >
                    {type.label}
                  </Text>
                  {docType === type.value && (
                    <Text style={styles.typeOptionCheck}>✓</Text>
                  )}
                </TouchableOpacity>
              ))}
            </View>
          </TouchableOpacity>
        </Modal>
      </View>
    );
  }

  if (mode === 'uploading') {
    return (
      <View style={styles.uploadingContainer}>
        <View style={styles.uploadingContent}>
          <Text style={styles.uploadingIcon}>📤</Text>
          <Text style={styles.uploadingTitle}>Uploading Document</Text>
          <Text style={styles.uploadingSubtitle}>{title}</Text>

          {uploadProgress && (
            <View style={styles.uploadingProgressContainer}>
              <View style={styles.uploadingProgressBar}>
                <View
                  style={[
                    styles.uploadingProgressFill,
                    { width: `${uploadProgress.percentage}%` },
                  ]}
                />
              </View>
              <Text style={styles.uploadingProgressText}>
                {uploadProgress.percentage}%
              </Text>
            </View>
          )}

          <Text style={styles.uploadingHint}>
            Please keep the app open while uploading
          </Text>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <CameraView
        ref={cameraRef}
        style={styles.camera}
        facing={facing}
        flash={flash}
      >
        <View style={styles.topControls}>
          <TouchableOpacity style={styles.controlButton} onPress={() => router.back()}>
            <Text style={styles.controlButtonText}>✕</Text>
          </TouchableOpacity>

          <View style={styles.topControlsCenter}>
            <TouchableOpacity
              style={[styles.controlButton, flash === 'on' && styles.controlButtonActive]}
              onPress={toggleFlash}
            >
              <Text style={styles.controlButtonText}>
                {flash === 'on' ? '⚡' : '⚡️'}
              </Text>
            </TouchableOpacity>
            <TouchableOpacity style={styles.controlButton} onPress={toggleFacing}>
              <Text style={styles.controlButtonText}>🔄</Text>
            </TouchableOpacity>
          </View>

          <View style={styles.controlSpacer} />
        </View>

        <View style={styles.bottomControls}>
          <View style={styles.hintContainer}>
            <Text style={styles.hintText}>
              Position your document within the frame
            </Text>
          </View>

          <View style={styles.captureRow}>
            <View style={styles.captureSpacer} />

            <Animated.View
              style={[
                styles.captureButtonOuter,
                { transform: [{ scale: captureAnim }] },
              ]}
            >
              <TouchableOpacity
                style={styles.captureButton}
                onPress={capturePhoto}
                disabled={mode === 'capturing'}
              >
                {mode === 'capturing' ? (
                  <ActivityIndicator size="small" color={tokens.accentColor} />
                ) : (
                  <View style={styles.captureButtonInner} />
                )}
              </TouchableOpacity>
            </Animated.View>

            <View style={styles.captureSpacer} />
          </View>

          <View style={styles.quickTypeContainer}>
            <Text style={styles.quickTypeLabel}>Quick select type:</Text>
            <ScrollView
              horizontal
              showsHorizontalScrollIndicator={false}
              contentContainerStyle={styles.quickTypeScroll}
            >
              {DOCUMENT_TYPES.map((type) => (
                <TouchableOpacity
                  key={type.value}
                  style={[
                    styles.quickTypeChip,
                    docType === type.value && styles.quickTypeChipActive,
                  ]}
                  onPress={() => setDocType(type.value)}
                >
                  <Text style={styles.quickTypeChipText}>
                    {type.icon} {type.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </ScrollView>
          </View>
        </View>
      </CameraView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: tokens.backgroundColor.primary,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: tokens.backgroundColor.primary,
  },
  loadingText: {
    color: tokens.textColor.secondary,
    fontSize: 16,
    marginTop: 16,
  },
  permissionContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: tokens.backgroundColor.primary,
    padding: 24,
  },
  permissionIcon: {
    fontSize: 64,
    marginBottom: 24,
  },
  permissionTitle: {
    color: tokens.textColor.primary,
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 12,
    textAlign: 'center',
  },
  permissionText: {
    color: tokens.textColor.secondary,
    fontSize: 16,
    textAlign: 'center',
    lineHeight: 24,
    marginBottom: 32,
  },
  permissionButton: {
    backgroundColor: tokens.accentColor,
    paddingHorizontal: 32,
    paddingVertical: 14,
    borderRadius: 12,
    marginBottom: 12,
    width: '100%',
    maxWidth: 280,
    alignItems: 'center',
  },
  permissionButtonText: {
    color: colors.gray[50],
    fontSize: 16,
    fontWeight: '600',
  },
  secondaryButton: {
    backgroundColor: colors.gray[800],
  },
  secondaryButtonText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  camera: {
    flex: 1,
  },
  topControls: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingTop: 60,
    paddingHorizontal: 20,
  },
  topControlsCenter: {
    flexDirection: 'row',
    gap: 12,
  },
  controlSpacer: {
    width: 44,
  },
  controlButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  controlButtonActive: {
    backgroundColor: tokens.accentColor,
  },
  controlButtonText: {
    fontSize: 20,
    color: colors.gray[50],
  },
  bottomControls: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    paddingBottom: 40,
  },
  hintContainer: {
    alignItems: 'center',
    marginBottom: 20,
  },
  hintText: {
    color: colors.gray[50],
    fontSize: 14,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
  },
  captureRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'center',
    marginBottom: 20,
  },
  captureSpacer: {
    width: 60,
  },
  captureButtonOuter: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  captureButton: {
    width: 68,
    height: 68,
    borderRadius: 34,
    backgroundColor: colors.gray[50],
    justifyContent: 'center',
    alignItems: 'center',
  },
  captureButtonInner: {
    width: 58,
    height: 58,
    borderRadius: 29,
    backgroundColor: colors.gray[50],
    borderWidth: 3,
    borderColor: tokens.accentColor,
  },
  quickTypeContainer: {
    paddingHorizontal: 16,
  },
  quickTypeLabel: {
    color: colors.gray[300],
    fontSize: 12,
    marginBottom: 8,
    textAlign: 'center',
  },
  quickTypeScroll: {
    paddingHorizontal: 8,
    gap: 8,
  },
  quickTypeChip: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    backgroundColor: 'rgba(255, 255, 255, 0.2)',
    marginHorizontal: 4,
  },
  quickTypeChipActive: {
    backgroundColor: tokens.accentColor,
  },
  quickTypeChipText: {
    color: colors.gray[50],
    fontSize: 13,
  },
  previewImage: {
    flex: 1,
    backgroundColor: tokens.backgroundColor.secondary,
  },
  formContainer: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    maxHeight: '50%',
    backgroundColor: colors.gray[900],
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
  },
  formContent: {
    padding: 24,
  },
  inputGroup: {
    marginBottom: 20,
  },
  inputLabel: {
    color: tokens.textColor.secondary,
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 8,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  textInput: {
    backgroundColor: colors.gray[800],
    borderRadius: 12,
    padding: 16,
    color: tokens.textColor.primary,
    fontSize: 16,
    borderWidth: 1,
    borderColor: colors.gray[700],
  },
  typeSelector: {
    backgroundColor: colors.gray[800],
    borderRadius: 12,
    padding: 16,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: colors.gray[700],
  },
  typeSelectorText: {
    color: tokens.textColor.primary,
    fontSize: 16,
  },
  typeSelectorArrow: {
    color: tokens.textColor.secondary,
    fontSize: 12,
  },
  errorContainer: {
    backgroundColor: `${colors.error[500]}22`,
    borderRadius: 8,
    padding: 12,
    marginTop: 8,
  },
  errorText: {
    color: colors.error[500],
    fontSize: 14,
    textAlign: 'center',
  },
  previewActions: {
    flexDirection: 'row',
    padding: 16,
    gap: 12,
    borderTopWidth: 1,
    borderTopColor: colors.gray[800],
  },
  retakeButton: {
    flex: 1,
    paddingVertical: 14,
    borderRadius: 8,
    backgroundColor: colors.gray[800],
    alignItems: 'center',
  },
  retakeButtonText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
  uploadButton: {
    flex: 2,
    paddingVertical: 14,
    borderRadius: 8,
    backgroundColor: tokens.accentColor,
    alignItems: 'center',
    justifyContent: 'center',
  },
  uploadButtonText: {
    color: colors.gray[50],
    fontSize: 16,
    fontWeight: '600',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: colors.gray[900],
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    paddingBottom: 40,
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 20,
    borderBottomWidth: 1,
    borderBottomColor: colors.gray[800],
  },
  modalTitle: {
    color: tokens.textColor.primary,
    fontSize: 20,
    fontWeight: 'bold',
  },
  modalClose: {
    color: tokens.textColor.secondary,
    fontSize: 24,
    padding: 4,
  },
  typeOption: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: colors.gray[800],
  },
  typeOptionActive: {
    backgroundColor: `${tokens.accentColor}22`,
  },
  typeOptionIcon: {
    fontSize: 24,
    marginRight: 16,
  },
  typeOptionText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    flex: 1,
  },
  typeOptionTextActive: {
    fontWeight: '600',
    color: tokens.accentColor,
  },
  typeOptionCheck: {
    color: tokens.accentColor,
    fontSize: 18,
    fontWeight: 'bold',
  },
  uploadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: tokens.backgroundColor.primary,
    padding: 24,
  },
  uploadingContent: {
    alignItems: 'center',
    width: '100%',
    maxWidth: 300,
  },
  uploadingIcon: {
    fontSize: 64,
    marginBottom: 24,
  },
  uploadingTitle: {
    color: tokens.textColor.primary,
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 8,
    textAlign: 'center',
  },
  uploadingSubtitle: {
    color: tokens.textColor.secondary,
    fontSize: 16,
    marginBottom: 32,
    textAlign: 'center',
  },
  uploadingProgressContainer: {
    width: '100%',
    marginBottom: 24,
  },
  uploadingProgressBar: {
    height: 12,
    backgroundColor: colors.gray[800],
    borderRadius: 6,
    overflow: 'hidden',
    marginBottom: 8,
  },
  uploadingProgressFill: {
    height: '100%',
    backgroundColor: tokens.accentColor,
    borderRadius: 6,
  },
  uploadingProgressText: {
    color: tokens.textColor.primary,
    fontSize: 18,
    fontWeight: '600',
    textAlign: 'center',
  },
  uploadingHint: {
    color: tokens.textColor.muted,
    fontSize: 14,
    textAlign: 'center',
  },
});
