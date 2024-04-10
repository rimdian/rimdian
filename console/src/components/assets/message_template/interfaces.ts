export interface MessageTemplate {
  id: string
  version: number
  name: string
  engine: 'visual' | 'code'
  channel: 'email'
  category: 'transactional' | 'campaign' | 'automation' | 'other'
  email: MessageTemplateEmail
  utm_source?: string
  utm_medium?: string
  utm_campaign?: string
  esp_settings: any
  test_data: any
  db_created_at: string
  db_updated_at: string
}

export interface MessageTemplateEmail {
  from_address: string
  from_name: string
  reply_to: string
  subject: string
  content: string
  text: string
}
