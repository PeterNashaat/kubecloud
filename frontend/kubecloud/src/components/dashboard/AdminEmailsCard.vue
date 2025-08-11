<template>
  <v-card class="pa-6" elevation="2">
    <v-card-title class="text-h4 pa-0 mb-2">
      Emails
    </v-card-title>
    <v-card-subtitle class="pa-0 mb-6">
      Send emails to all platform users
    </v-card-subtitle>
    
    <v-card-text class="pa-0">
      <v-form @submit.prevent="sendEmail">
        <v-text-field
          v-model="title"
          label="Email Subject"
          placeholder="System Maintenance"
          variant="outlined"
          class="mb-4"
          :rules="[v => !!v || 'Subject is required']"
        ></v-text-field>
        
        <v-select
          v-model="priority"
          label="Priority Level"
          :items="priorityLevels"
          variant="outlined"
          class="mb-4"
        ></v-select>
        
        <v-textarea
          v-model="message"
          label="Email Content"
          placeholder="Compose your email message here..."
          variant="outlined"
          rows="6"
          class="mb-4"
          :rules="[v => !!v || 'Content is required']"
        ></v-textarea>
        
        <v-file-input
          v-model="attachments"
          label="Attachments (optional)"
          variant="outlined"
          multiple
          show-size
          counter
          class="mb-4"
          accept=".pdf,.doc,.docx,.txt,.jpg,.jpeg,.png,.gif,.zip"
          :rules="attachmentRules"
        >
          <template v-slot:selection="{ fileNames }">
            <template v-for="(fileName, index) in fileNames" :key="fileName">
              <v-chip
                v-if="index < 2"
                color="primary"
                size="small"
                class="me-2"
              >
                {{ fileName }}
              </v-chip>
              <span
                v-else-if="index === 2"
                class="text-overline grey--text text--darken-3 mx-2"
              >
                +{{ fileNames.length - 2 }} File(s)
              </span>
            </template>
          </template>
        </v-file-input>
        
        <v-btn
          type="submit"
          color="primary"
          :loading="sending"
          :disabled="!title || !message || sending"
          :size="$vuetify.display.xs ? 'default' : 'large'"
          :block="$vuetify.display.xs"
          class="mt-4"
        >
          <span class="d-none d-sm-inline">Send to All Users</span>
          <span class="d-sm-none">Send Email</span>
        </v-btn>
        
        <v-alert
          v-if="result"
          :type="result.success ? 'success' : 'error'"
          class="mt-4"
          closable
          variant="tonal"
        >
          {{ result.message }}
        </v-alert>
      </v-form>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { adminService } from '@/utils/adminService'

// Form state
const title = ref('')
const message = ref('')
const priority = ref('Normal')
const attachments = ref<File[]>([])
const sending = ref(false)
const result = ref<{ success: boolean; message: string } | null>(null)

const priorityLevels = ['Low', 'Normal', 'High', 'Urgent']

// File validation rules
const attachmentRules = [
  (files: File[]) => {
    if (!files || files.length === 0) return true
    const maxSize = 10 * 1024 * 1024 // 10MB
    const oversized = files.find(file => file.size > maxSize)
    if (oversized) return `File "${oversized.name}" is too large. Maximum size is 10MB.`
    return true
  },
  (files: File[]) => {
    if (!files || files.length === 0) return true
    if (files.length > 5) return 'Maximum 5 files allowed'
    return true
  }
]


// Computed properties for the preview
const priorityIcon = computed(() => {
  switch (priority.value.toLowerCase()) {
    case 'low': return 'mdi-information-outline'
    case 'high': return 'mdi-alert-outline'
    case 'urgent': return 'mdi-alert-octagon-outline'
    default: return 'mdi-email-outline'
  }
})

const priorityColor = computed(() => {
  switch (priority.value.toLowerCase()) {
    case 'low': return 'blue-lighten-1'
    case 'high': return 'amber-darken-1'
    case 'urgent': return 'red-darken-1'
    default: return 'blue-grey-lighten-1'
  }
})

const currentDateFormatted = computed(() => {
  return new Date().toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
    hour12: true
  })
})

// Methods
async function sendEmail() {
  if (!title.value || !message.value) return

  sending.value = true
  try {
    // Prepare form data for file upload
    const formData = new FormData()
    formData.append('title', title.value)
    formData.append('message', message.value)
    formData.append('priority', priority.value)
    
    // Add attachments if any
    if (attachments.value && attachments.value.length > 0) {
      attachments.value.forEach((file, index) => {
        formData.append(`attachments`, file)
      })
    }

    // Call the API to send email with attachments
    const response = await adminService.sendSystemEmailWithAttachments(formData)

    result.value = {
      success: true,
      message: `Email has been sent to all users successfully${attachments.value.length > 0 ? ` with ${attachments.value.length} attachment(s)` : ''}.`
    }

    // Reset form
    title.value = ''
    message.value = ''
    priority.value = 'Normal'
    attachments.value = []
  } catch (error) {
    result.value = {
      success: false,
      message: 'Failed to send email. Please try again.'
    }
  } finally {
    sending.value = false
  }
}

function formatDate(dateString: string) {
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
    hour12: true
  })
}

function formatDateMobile(dateString: string) {
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
    hour12: true
  })
}

function getPriorityColor(priority: string) {
  switch (priority.toLowerCase()) {
    case 'urgent': return 'red'
    case 'high': return 'orange'
    case 'low': return 'blue-lighten-1'
    default: return 'primary'
  }
}

</script>

