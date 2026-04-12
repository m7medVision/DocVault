import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  ScrollView,
  TextInput,
  Modal,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { tokens, colors } from '../../src/theme/tokens';
import { useUpload, DOCUMENT_TYPES } from '../../src/features/documents/useUpload';
import { formatFileSize, getFileTypeIcon } from '../../src/features/documents/helpers';

export default function FilePickerUploadScreen() {
  const router = useRouter();
  const { accessToken } = useAuth();

  const {
    mode,
    selectedFile,
    uploadProgress,
    uploadError,
    uploadResult,
    title,
    docType,
    titleError,
    setTitle,
    setDocType,
    setTitleError,
    pickDocument,
    clearSelection,
    uploadDocument,
    retryUpload,
  } = useUpload(accessToken);

  const goToDocuments = () => {
    router.replace('/(app)/documents');
  };

  const renderFilePreview = () => {
    if (!selectedFile) return null;
    return (
      <View style={styles.filePreview}>
        <View style={styles.fileIconContainer}>
          <Text style={styles.fileIcon}>{getFileTypeIcon(selectedFile.mimeType)}</Text>
        </View>
        <View style={styles.fileInfo}>
          <Text style={styles.fileName} numberOfLines={2}>
            {selectedFile.name}
          </Text>
          <View style={styles.fileMeta}>
            <Text style={styles.fileMetaText}>
              {selectedFile.type.toUpperCase()} • {formatFileSize(selectedFile.size)}
            </Text>
          </View>
        </View>
        <TouchableOpacity style={styles.removeButton} onPress={clearSelection}>
          <Text style={styles.removeButtonText}>✕</Text>
        </TouchableOpacity>
      </View>
    );
  };

  const renderSelectMode = () => (
    <View style={styles.selectContainer}>
      <View style={styles.selectContent}>
        <Text style={styles.selectIcon}>📁</Text>
        <Text style={styles.selectTitle}>Upload Document</Text>
        <Text style={styles.selectSubtitle}>
          Select a PDF or image from your device to upload
        </Text>
        
        <View style={styles.supportedFormats}>
          <Text style={styles.supportedFormatsTitle}>Supported formats:</Text>
          <View style={styles.formatChips}>
            {['PDF', 'JPEG', 'PNG', 'HEIC'].map((format) => (
              <View key={format} style={styles.formatChip}>
                <Text style={styles.formatChipText}>{format}</Text>
              </View>
            ))}
          </View>
        </View>

        <TouchableOpacity style={styles.selectButton} onPress={pickDocument}>
          <Text style={styles.selectButtonText}>Choose File</Text>
        </TouchableOpacity>

        <Text style={styles.sizeLimitText}>Maximum file size: 50MB</Text>
      </View>

      <View style={styles.quickActions}>
        <Text style={styles.quickActionsTitle}>Quick Upload</Text>
        <View style={styles.quickActionsRow}>
          <TouchableOpacity
            style={styles.quickActionButton}
            onPress={() => router.push('/(app)/camera')}
          >
            <Text style={styles.quickActionIcon}>📷</Text>
            <Text style={styles.quickActionLabel}>Camera</Text>
          </TouchableOpacity>
        </View>
      </View>
    </View>
  );

  const renderPreviewMode = () => (
    <KeyboardAvoidingView
      style={styles.previewContainer}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.previewScrollContent}>
        {renderFilePreview()}

        <View style={styles.formContainer}>
          <View style={styles.inputGroup}>
            <Text style={styles.inputLabel}>Document Title</Text>
            <TextInput
              style={[styles.textInput, titleError && styles.textInputError]}
              value={title}
              onChangeText={(text) => {
                setTitle(text);
                if (titleError) setTitleError(null);
              }}
              placeholder="Enter document title"
              placeholderTextColor={colors.gray[500]}
              autoCapitalize="words"
              maxLength={200}
            />
            {titleError && (
              <Text style={styles.errorText}>{titleError}</Text>
            )}
            <Text style={styles.charCount}>
              {title.length}/200
            </Text>
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.inputLabel}>Document Type</Text>
            <TouchableOpacity
              style={styles.typeSelector}
              onPress={() => {}}
            >
              <Text style={styles.typeSelectorText}>
                {DOCUMENT_TYPES.find((t) => t.value === docType)?.icon}{' '}
                {DOCUMENT_TYPES.find((t) => t.value === docType)?.label}
              </Text>
              <Text style={styles.typeSelectorArrow}>▼</Text>
            </TouchableOpacity>
          </View>
        </View>
      </ScrollView>

      <View style={styles.previewActions}>
        <TouchableOpacity style={styles.cancelButton} onPress={clearSelection}>
          <Text style={styles.cancelButtonText}>Cancel</Text>
        </TouchableOpacity>
        <TouchableOpacity style={styles.uploadButton} onPress={uploadDocument}>
          <Text style={styles.uploadButtonText}>Upload Document</Text>
        </TouchableOpacity>
      </View>

      <Modal visible={false} transparent animationType="slide">
        <TouchableOpacity style={styles.modalOverlay} activeOpacity={1}>
          <View style={styles.modalContent}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>Select Document Type</Text>
              <TouchableOpacity>
                <Text style={styles.modalClose}>✕</Text>
              </TouchableOpacity>
            </View>
            <ScrollView>
              {DOCUMENT_TYPES.map((type) => (
                <TouchableOpacity
                  key={type.value}
                  style={[
                    styles.typeOption,
                    docType === type.value && styles.typeOptionActive,
                  ]}
                  onPress={() => {
                    setDocType(type.value);
                  }}
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
            </ScrollView>
          </View>
        </TouchableOpacity>
      </Modal>
    </KeyboardAvoidingView>
  );

  const renderUploadingMode = () => (
    <View style={styles.uploadingContainer}>
      <View style={styles.uploadingContent}>
        <Text style={styles.uploadingIcon}>📤</Text>
        <Text style={styles.uploadingTitle}>Uploading Document</Text>
        <Text style={styles.uploadingSubtitle}>{title}</Text>

        {uploadProgress && (
          <View style={styles.progressContainer}>
            <View style={styles.progressBar}>
              <View
                style={[
                  styles.progressFill,
                  { width: `${uploadProgress.percentage}%` },
                ]}
              />
            </View>
            <View style={styles.progressStats}>
              <Text style={styles.progressPercentage}>
                {uploadProgress.percentage}%
              </Text>
              <Text style={styles.progressBytes}>
                {formatFileSize(uploadProgress.bytesUploaded)} / {formatFileSize(uploadProgress.totalBytes)}
              </Text>
            </View>
          </View>
        )}

        <Text style={styles.uploadingHint}>
          Please keep the app open while uploading
        </Text>
      </View>
    </View>
  );

  const renderSuccessMode = () => (
    <View style={styles.resultContainer}>
      <View style={styles.resultContent}>
        <Text style={styles.resultIcon}>✅</Text>
        <Text style={styles.resultTitle}>Upload Complete!</Text>
        <Text style={styles.resultSubtitle}>
          Document "{title}" has been uploaded successfully and is being processed.
        </Text>
        
        {uploadResult && (
          <View style={styles.resultDetails}>
            <Text style={styles.resultDetailText}>
              Document ID: {uploadResult.id}
            </Text>
            <Text style={styles.resultDetailText}>
              Status: {uploadResult.status}
            </Text>
          </View>
        )}

        <TouchableOpacity style={styles.resultButton} onPress={goToDocuments}>
          <Text style={styles.resultButtonText}>Go to Documents</Text>
        </TouchableOpacity>

        <TouchableOpacity style={styles.uploadAnotherButton} onPress={clearSelection}>
          <Text style={styles.uploadAnotherButtonText}>Upload Another</Text>
        </TouchableOpacity>
      </View>
    </View>
  );

  const renderErrorMode = () => (
    <View style={styles.resultContainer}>
      <View style={styles.resultContent}>
        <Text style={styles.resultIcon}>❌</Text>
        <Text style={styles.resultTitle}>Upload Failed</Text>
        <Text style={styles.resultSubtitle}>
          {uploadError || 'An error occurred while uploading your document.'}
        </Text>

        <View style={styles.errorActions}>
          <TouchableOpacity style={styles.retryButton} onPress={retryUpload}>
            <Text style={styles.retryButtonText}>Try Again</Text>
          </TouchableOpacity>

          <TouchableOpacity style={styles.goBackButton} onPress={clearSelection}>
            <Text style={styles.goBackButtonText}>Choose Different File</Text>
          </TouchableOpacity>
        </View>
      </View>
    </View>
  );

  const renderContent = () => {
    switch (mode) {
      case 'select':
        return renderSelectMode();
      case 'preview':
        return renderPreviewMode();
      case 'uploading':
        return renderUploadingMode();
      case 'success':
        return renderSuccessMode();
      case 'error':
        return renderErrorMode();
      default:
        return renderSelectMode();
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity style={styles.backButton} onPress={() => router.back()}>
          <Text style={styles.backButtonText}>←</Text>
        </TouchableOpacity>
        <Text style={styles.headerTitle}>Upload Document</Text>
        <View style={styles.headerSpacer} />
      </View>

      {renderContent()}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: tokens.backgroundColor.primary,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingTop: 60,
    paddingHorizontal: 16,
    paddingBottom: 16,
    backgroundColor: tokens.backgroundColor.primary,
    borderBottomWidth: 1,
    borderBottomColor: colors.gray[800],
  },
  backButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: colors.gray[800],
    justifyContent: 'center',
    alignItems: 'center',
  },
  backButtonText: {
    fontSize: 24,
    color: tokens.textColor.primary,
  },
  headerTitle: {
    color: tokens.textColor.primary,
    fontSize: 18,
    fontWeight: '600',
  },
  headerSpacer: {
    width: 44,
  },
  selectContainer: {
    flex: 1,
    padding: 24,
  },
  selectContent: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  selectIcon: {
    fontSize: 64,
    marginBottom: 24,
  },
  selectTitle: {
    color: tokens.textColor.primary,
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 8,
    textAlign: 'center',
  },
  selectSubtitle: {
    color: tokens.textColor.secondary,
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 32,
    lineHeight: 24,
  },
  supportedFormats: {
    alignItems: 'center',
    marginBottom: 32,
  },
  supportedFormatsTitle: {
    color: tokens.textColor.muted,
    fontSize: 12,
    marginBottom: 12,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  formatChips: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    gap: 8,
  },
  formatChip: {
    backgroundColor: colors.gray[800],
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
  },
  formatChipText: {
    color: tokens.textColor.secondary,
    fontSize: 13,
    fontWeight: '500',
  },
  selectButton: {
    backgroundColor: tokens.accentColor,
    paddingHorizontal: 48,
    paddingVertical: 16,
    borderRadius: 12,
    marginBottom: 16,
  },
  selectButtonText: {
    color: colors.gray[50],
    fontSize: 16,
    fontWeight: '600',
  },
  sizeLimitText: {
    color: tokens.textColor.muted,
    fontSize: 12,
  },
  quickActions: {
    paddingTop: 24,
    borderTopWidth: 1,
    borderTopColor: colors.gray[800],
  },
  quickActionsTitle: {
    color: tokens.textColor.muted,
    fontSize: 12,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: 16,
    textAlign: 'center',
  },
  quickActionsRow: {
    flexDirection: 'row',
    justifyContent: 'center',
  },
  quickActionButton: {
    alignItems: 'center',
    padding: 16,
  },
  quickActionIcon: {
    fontSize: 32,
    marginBottom: 8,
  },
  quickActionLabel: {
    color: tokens.textColor.secondary,
    fontSize: 14,
  },
  previewContainer: {
    flex: 1,
  },
  previewScrollContent: {
    flexGrow: 1,
    padding: 24,
  },
  filePreview: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.gray[800],
    borderRadius: 12,
    padding: 16,
    marginBottom: 24,
    borderWidth: 1,
    borderColor: colors.gray[700],
  },
  fileIconContainer: {
    width: 56,
    height: 56,
    borderRadius: 12,
    backgroundColor: `${tokens.accentColor}22`,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  fileIcon: {
    fontSize: 28,
  },
  fileInfo: {
    flex: 1,
  },
  fileName: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 4,
  },
  fileMeta: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  fileMetaText: {
    color: tokens.textColor.secondary,
    fontSize: 13,
  },
  removeButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: colors.gray[700],
    justifyContent: 'center',
    alignItems: 'center',
  },
  removeButtonText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
    fontWeight: 'bold',
  },
  formContainer: {
    flex: 1,
  },
  inputGroup: {
    marginBottom: 24,
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
  textInputError: {
    borderColor: colors.error[500],
  },
  errorText: {
    color: colors.error[500],
    fontSize: 12,
    marginTop: 6,
  },
  charCount: {
    color: tokens.textColor.muted,
    fontSize: 11,
    textAlign: 'right',
    marginTop: 4,
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
  previewActions: {
    flexDirection: 'row',
    padding: 16,
    gap: 12,
    borderTopWidth: 1,
    borderTopColor: colors.gray[800],
    backgroundColor: tokens.backgroundColor.primary,
  },
  cancelButton: {
    flex: 1,
    paddingVertical: 14,
    borderRadius: 8,
    backgroundColor: colors.gray[800],
    alignItems: 'center',
  },
  cancelButtonText: {
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
    maxHeight: '70%',
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
  progressContainer: {
    width: '100%',
    marginBottom: 24,
  },
  progressBar: {
    height: 12,
    backgroundColor: colors.gray[800],
    borderRadius: 6,
    overflow: 'hidden',
    marginBottom: 12,
  },
  progressFill: {
    height: '100%',
    backgroundColor: tokens.accentColor,
    borderRadius: 6,
  },
  progressStats: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  progressPercentage: {
    color: tokens.textColor.primary,
    fontSize: 18,
    fontWeight: '600',
  },
  progressBytes: {
    color: tokens.textColor.secondary,
    fontSize: 14,
  },
  uploadingHint: {
    color: tokens.textColor.muted,
    fontSize: 14,
    textAlign: 'center',
  },
  resultContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  resultContent: {
    alignItems: 'center',
    width: '100%',
    maxWidth: 320,
  },
  resultIcon: {
    fontSize: 64,
    marginBottom: 24,
  },
  resultTitle: {
    color: tokens.textColor.primary,
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 8,
    textAlign: 'center',
  },
  resultSubtitle: {
    color: tokens.textColor.secondary,
    fontSize: 16,
    marginBottom: 24,
    textAlign: 'center',
    lineHeight: 24,
  },
  resultDetails: {
    backgroundColor: colors.gray[800],
    borderRadius: 12,
    padding: 16,
    width: '100%',
    marginBottom: 24,
  },
  resultDetailText: {
    color: tokens.textColor.secondary,
    fontSize: 13,
    fontFamily: 'monospace',
    marginBottom: 4,
  },
  resultButton: {
    backgroundColor: tokens.accentColor,
    paddingHorizontal: 32,
    paddingVertical: 14,
    borderRadius: 8,
    width: '100%',
    alignItems: 'center',
    marginBottom: 12,
  },
  resultButtonText: {
    color: colors.gray[50],
    fontSize: 16,
    fontWeight: '600',
  },
  uploadAnotherButton: {
    paddingHorizontal: 32,
    paddingVertical: 14,
    width: '100%',
    alignItems: 'center',
  },
  uploadAnotherButtonText: {
    color: tokens.textColor.secondary,
    fontSize: 16,
  },
  errorActions: {
    width: '100%',
    gap: 12,
  },
  retryButton: {
    backgroundColor: tokens.accentColor,
    paddingHorizontal: 32,
    paddingVertical: 14,
    borderRadius: 8,
    width: '100%',
    alignItems: 'center',
    marginBottom: 12,
  },
  retryButtonText: {
    color: colors.gray[50],
    fontSize: 16,
    fontWeight: '600',
  },
  goBackButton: {
    backgroundColor: colors.gray[800],
    paddingHorizontal: 32,
    paddingVertical: 14,
    borderRadius: 8,
    width: '100%',
    alignItems: 'center',
  },
  goBackButtonText: {
    color: tokens.textColor.primary,
    fontSize: 16,
    fontWeight: '600',
  },
});
