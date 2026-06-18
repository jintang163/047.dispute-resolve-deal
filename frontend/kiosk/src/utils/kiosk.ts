import QRCode from 'qrcode'
import type { IdCardInfo } from '@/stores/kiosk'

export interface CardReaderDevice {
  id: string
  name: string
  status: 'connected' | 'disconnected' | 'error'
}

export interface CardReaderResult {
  success: boolean
  data?: IdCardInfo
  error?: string
}

declare global {
  interface Window {
    CardReaderAPI?: {
      listDevices: () => Promise<CardReaderDevice[]>
      readCard: (deviceId?: string, timeout?: number) => Promise<CardReaderResult>
      stopRead: () => void
      onDeviceChange?: (devices: CardReaderDevice[]) => void
    }
    PrinterAPI?: {
      listPrinters: () => Promise<string[]>
      printReceipt: (printerName: string, content: ReceiptContent) => Promise<boolean>
      printTest: (printerName: string) => Promise<boolean>
    }
  }
}

export interface ReceiptContent {
  caseNumber: string
  title: string
  items: { label: string; value: string }[]
  qrCodeData: string
  footerText?: string
}

export async function listCardReaders(): Promise<CardReaderDevice[]> {
  if (window.CardReaderAPI?.listDevices) {
    try {
      return await window.CardReaderAPI.listDevices()
    } catch (e) {
      console.warn('Card reader API error:', e)
    }
  }
  return []
}

export async function readIdCard(deviceId?: string, timeout: number = 15000): Promise<CardReaderResult> {
  if (window.CardReaderAPI?.readCard) {
    try {
      return await window.CardReaderAPI.readCard(deviceId, timeout)
    } catch (e) {
      return {
        success: false,
        error: e instanceof Error ? e.message : '读取失败'
      }
    }
  }
  return {
    success: false,
    error: '读卡器API不可用，请手动填写'
  }
}

export function stopCardReading(): void {
  if (window.CardReaderAPI?.stopRead) {
    window.CardReaderAPI.stopRead()
  }
}

export async function generateQRCode(
  data: string,
  options: {
    width?: number
    margin?: number
    color?: { dark?: string; light?: string }
  } = {}
): Promise<string> {
  try {
    return await QRCode.toDataURL(data, {
      width: options.width || 320,
      margin: options.margin || 2,
      color: {
        dark: options.color?.dark || '#1a1a1a',
        light: options.color?.light || '#ffffff'
      },
      errorCorrectionLevel: 'M'
    })
  } catch (e) {
    console.error('QR code generation failed:', e)
    return ''
  }
}

export async function generateQRCodeToCanvas(
  canvas: HTMLCanvasElement,
  data: string,
  options: { width?: number } = {}
): Promise<void> {
  try {
    await QRCode.toCanvas(canvas, data, {
      width: options.width || 320,
      errorCorrectionLevel: 'M'
    })
  } catch (e) {
    console.error('QR code canvas generation failed:', e)
  }
}

export async function listPrinters(): Promise<string[]> {
  if (window.PrinterAPI?.listPrinters) {
    try {
      return await window.PrinterAPI.listPrinters()
    } catch (e) {
      console.warn('Printer API error:', e)
    }
  }
  return []
}

export async function printReceipt(content: ReceiptContent, printerName?: string): Promise<boolean> {
  if (window.PrinterAPI?.printReceipt) {
    try {
      const name = printerName || (await listPrinters())[0]
      if (!name) return false
      return await window.PrinterAPI.printReceipt(name, content)
    } catch (e) {
      console.error('Print failed:', e)
      return false
    }
  }
  console.log('[Mock Print] Receipt:', content)
  return true
}

export async function printTestPage(printerName?: string): Promise<boolean> {
  if (window.PrinterAPI?.printTest) {
    try {
      const name = printerName || (await listPrinters())[0]
      if (!name) return false
      return await window.PrinterAPI.printTest(name)
    } catch (e) {
      console.error('Test print failed:', e)
      return false
    }
  }
  console.log('[Mock Print] Test page')
  return true
}

export function generateCaseNumber(): string {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const random = Math.floor(Math.random() * 10000).toString().padStart(4, '0')
  return `JF${year}${month}${day}${random}`
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

export function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2, 9)
}

export function validateIdNumber(idNumber: string): boolean {
  const reg = /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/
  if (!reg.test(idNumber)) return false
  const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
  const checkCodes = ['1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2']
  let sum = 0
  for (let i = 0; i < 17; i++) {
    sum += parseInt(idNumber[i]) * weights[i]
  }
  return checkCodes[sum % 11] === idNumber[17].toUpperCase()
}

export function validatePhone(phone: string): boolean {
  return /^1[3-9]\d{9}$/.test(phone)
}
