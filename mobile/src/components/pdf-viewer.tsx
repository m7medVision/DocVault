import React, { useCallback, useRef, useState } from 'react';
import { StyleSheet, View, Platform } from 'react-native';
import { WebView } from 'react-native-webview';

import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';

const PDFJS_CDN = 'https://cdn.jsdelivr.net/npm/pdfjs-dist@5.4.296/build';

function generatePdfHtml(url: string): string {
  return `<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=3">
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { background: #f4f4f5; font-family: -apple-system, sans-serif; }
#toolbar { display: flex; align-items: center; justify-content: center; gap: 16px; padding: 12px; background: #fff; border-bottom: 1px solid #e4e4e7; position: sticky; top: 0; z-index: 10; }
#toolbar button { padding: 8px 16px; border: 1px solid #d4d4d8; border-radius: 8px; background: #fff; font-size: 14px; cursor: pointer; }
#toolbar button:disabled { opacity: 0.3; }
#pageInfo { font-size: 13px; color: #71717a; min-width: 80px; text-align: center; }
#container { display: flex; justify-content: center; padding: 12px; }
#canvasWrapper { position: relative; display: inline-block; }
canvas { display: block; border-radius: 4px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
#loading { text-align: center; padding: 40px; color: #71717a; font-size: 14px; }
#error { text-align: center; padding: 40px; color: #dc2626; font-size: 14px; }
.spinner { display: inline-block; width: 24px; height: 24px; border: 3px solid #e4e4e7; border-top-color: #208AEF; border-radius: 50%; animation: spin 0.8s linear infinite; margin-bottom: 12px; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
</head>
<body>
<div id="toolbar">
  <button id="prevBtn" disabled>Prev</button>
  <span id="pageInfo">-- / --</span>
  <button id="nextBtn" disabled>Next</button>
</div>
<div id="container">
  <div id="canvasWrapper"></div>
</div>
<div id="loading">
  <div class="spinner"></div>
  <div>Loading PDF...</div>
</div>
<div id="error" style="display:none;">Failed to load PDF</div>
<script type="module">
import * as pdfjsLib from '${PDFJS_CDN}/pdf.min.mjs';
pdfjsLib.GlobalWorkerOptions.workerSrc = '${PDFJS_CDN}/pdf.worker.min.mjs';

const url = ${JSON.stringify(url)};
let pdfDoc = null;
let currentPage = 1;
let totalPages = 0;
let pageRendering = false;

const prevBtn = document.getElementById('prevBtn');
const nextBtn = document.getElementById('nextBtn');
const pageInfo = document.getElementById('pageInfo');
const container = document.getElementById('container');
const canvasWrapper = document.getElementById('canvasWrapper');
const loadingEl = document.getElementById('loading');
const errorEl = document.getElementById('error');

function sendMsg(type, data) {
  window.ReactNativeWebView?.postMessage(JSON.stringify({ type, ...data }));
}

function updateUI() {
  prevBtn.disabled = currentPage <= 1;
  nextBtn.disabled = currentPage >= totalPages;
  pageInfo.textContent = currentPage + ' / ' + totalPages;
}

async function renderPage(num) {
  if (!pdfDoc) return;
  pageRendering = true;
  try {
    const page = await pdfDoc.getPage(num);
    const viewport = page.getViewport({ scale: 1.5 });
    const containerWidth = container.clientWidth - 24;
    const scale = containerWidth / viewport.width;
    const scaledViewport = page.getViewport({ scale: scale });

    const canvas = document.createElement('canvas');
    canvas.width = scaledViewport.width;
    canvas.height = scaledViewport.height;
    const ctx = canvas.getContext('2d');
    await page.render({ canvasContext: ctx, viewport: scaledViewport }).promise;

    canvasWrapper.innerHTML = '';
    canvasWrapper.appendChild(canvas);
  } catch (e) {
    console.error('Page render error:', e);
  } finally {
    pageRendering = false;
  }
}

function queueRender(num) {
  if (pageRendering) return;
  currentPage = num;
  updateUI();
  renderPage(num);
}

prevBtn.addEventListener('click', () => {
  if (currentPage > 1) queueRender(currentPage - 1);
});
nextBtn.addEventListener('click', () => {
  if (currentPage < totalPages) queueRender(currentPage + 1);
});

document.addEventListener('DOMContentLoaded', async () => {
  try {
    pdfDoc = await pdfjsLib.getDocument(url).promise;
    totalPages = pdfDoc.numPages;
    updateUI();
    loadingEl.style.display = 'none';
    queueRender(1);
  } catch (e) {
    loadingEl.style.display = 'none';
    errorEl.style.display = 'block';
    sendMsg('error', { message: e.message || 'Failed to load PDF' });
  }
});
</script>
</body>
</html>`;
}

interface PdfViewerProps {
  url: string;
}

export function PdfViewer({ url }: PdfViewerProps) {
  const theme = useTheme();
  const webViewRef = useRef<WebView>(null);
  const [error, setError] = useState<string | null>(null);

  const html = generatePdfHtml(url);
  const baseUrl = (() => {
    try {
      return new URL(url).origin;
    } catch {
      return undefined;
    }
  })();

  const handleMessage = useCallback((event: any) => {
    try {
      const data = JSON.parse(event.nativeEvent.data);
      if (data.type === 'error') {
        setError(data.message);
      }
    } catch {}
  }, []);

  if (error) {
    return (
      <View style={[styles.center, { backgroundColor: theme.surface }]}>
        <ThemedText type="small" themeColor="danger">{error}</ThemedText>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <WebView
        ref={webViewRef}
        source={{ html, baseUrl }}
        style={styles.webview}
        originWhitelist={['*']}
        onMessage={handleMessage}
        javaScriptEnabled
        domStorageEnabled
        allowFileAccess
        allowUniversalAccessFromFileURLs
        mixedContentMode="always"
        scalesPageToFit={Platform.OS === 'android'}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  webview: {
    flex: 1,
  },
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: Spacing.four,
    borderRadius: Spacing.two,
  },
});
