import { Copy, Eye, EyeOff, AlertTriangle, Info, Target, MapPin, Lightbulb } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useApiKeysContext } from '../context/apikeys-context'

export function ApiKeysViewDialog() {
  const { t } = useTranslation()
  const { isDialogOpen, closeDialog, selectedApiKey } = useApiKeysContext()
  const [isVisible, setIsVisible] = useState(false)

  const copyToClipboard = () => {
    if (selectedApiKey?.key) {
      navigator.clipboard.writeText(selectedApiKey.key)
      toast.success(t('apikeys.messages.copied'))
    }
  }

  const maskedKey = selectedApiKey?.key
    ? selectedApiKey.key.replace(/./g, '*').slice(0, -4) + selectedApiKey.key.slice(-4)
    : ''

  return (
    <Dialog open={isDialogOpen.view} onOpenChange={() => closeDialog()}>
      <DialogContent className="flex max-h-[90vh] flex-col sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t('apikeys.dialogs.view.title')}</DialogTitle>
          <DialogDescription>
            {t('apikeys.dialogs.view.description')}
          </DialogDescription>
        </DialogHeader>
        
        <Alert className="border-orange-200 bg-orange-50 dark:border-orange-800 dark:bg-orange-950">
          <AlertTriangle className="h-4 w-4 text-orange-600 dark:text-orange-400" />
          <AlertDescription className="text-orange-800 dark:text-orange-200">
            {t('apikeys.dialogs.view.warning')}
          </AlertDescription>
        </Alert>

        <div className="space-y-4 overflow-auto flex-1">
          <div>
            <label className="text-sm font-medium">{t('apikeys.columns.name')}</label>
            <div className="mt-1 p-3 bg-muted rounded-md">
              {selectedApiKey?.name}
            </div>
          </div>
          
          <div>
            <label className="text-sm font-medium">{t('apikeys.columns.key')}</label>
            <div className="mt-1 flex items-center space-x-2">
              <code className="flex-1 p-3 bg-muted rounded-md font-mono text-sm break-all">
                {isVisible ? selectedApiKey?.key : maskedKey}
              </code>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsVisible(!isVisible)}
                className="flex-shrink-0"
              >
                {isVisible ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={copyToClipboard}
                className="flex-shrink-0"
              >
                <Copy className="h-4 w-4" />
              </Button>
            </div>
          </div>

          <Alert className="border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-950">
            <Info className="h-4 w-4 text-blue-600 dark:text-blue-400" />
            <AlertDescription className="text-blue-800 dark:text-blue-200">
              <div className="space-y-4">
                {/* Title and Description */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Target className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                    <p className="font-semibold">{t('apikeys.dialogs.view.channelUsage.title')}</p>
                  </div>
                  <p className="text-sm">{t('apikeys.dialogs.view.channelUsage.description')}</p>
                </div>

                {/* How to Find ID Section */}
                <div className="space-y-2 pt-2 border-t border-blue-200 dark:border-blue-800">
                  <div className="flex items-center gap-2">
                    <MapPin className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                    <p className="font-semibold text-sm">{t('apikeys.dialogs.view.channelUsage.howToFindId')}</p>
                  </div>
                  <ol className="text-sm space-y-1.5 ml-6 list-decimal">
                    <li>{t('apikeys.dialogs.view.channelUsage.step1')}</li>
                    <li>{t('apikeys.dialogs.view.channelUsage.step2')}</li>
                    <li>{t('apikeys.dialogs.view.channelUsage.step3')}</li>
                  </ol>
                </div>

                {/* Example Section */}
                <div className="space-y-2 pt-2 border-t border-blue-200 dark:border-blue-800">
                  <div className="flex items-center gap-2">
                    <Lightbulb className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                    <p className="font-semibold text-sm">{t('apikeys.dialogs.view.channelUsage.exampleTitle')}</p>
                  </div>
                  <div className="p-3 bg-blue-100/50 dark:bg-blue-900/30 rounded-md">
                    <code className="text-sm break-all font-sans">
                      {t('apikeys.dialogs.view.channelUsage.example', { key: selectedApiKey?.key || 'your-api-key' })}
                    </code>
                  </div>
                </div>
              </div>
            </AlertDescription>
          </Alert>
        </div>
      </DialogContent>
    </Dialog>
  )
}