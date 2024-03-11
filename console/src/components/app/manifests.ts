import { AppManifest } from 'interfaces'

const googleCM360: AppManifest = {
  id: 'appx_googlecm360',
  name: 'Google Campaign Manager 360',
  homepage: 'https://www.rimdian.com/',
  author: 'Rimdian',
  icon_url: 'https://eu.rimdian.com/images/apps/googlecm360.png',
  short_description: 'Import postview conversions from Google Campaign Manager 360',
  description: 'Import postview conversions from Google Campaign Manager 360.',
  version: '1.0.0',
  ui_endpoint: 'https://nativeapps.rimdian.com',
  webhook_endpoint: 'https://nativeapps.rimdian.com/api/webhooks',
  tasks: [
    {
      id: 'appx_googlecm360_postview',
      name: 'Import new Google CM 360 postview data',
      is_cron: true,
      minutes_interval: 360,
      on_multiple_exec: 'discard_new'
    }
  ],
  extra_columns: [
    {
      kind: 'postview',
      columns: [
        {
          name: 'appx_googlecm360_campaign_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Campaign ID.'
        },
        {
          name: 'appx_googlecm360_campaign_name',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Campaign name.'
        },
        {
          name: 'appx_googlecm360_ad_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Ad ID.'
        },
        {
          name: 'appx_googlecm360_ad_name',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Ad name.'
        },
        {
          name: 'appx_googlecm360_ad_size',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Ad size.'
        },
        {
          name: 'appx_googlecm360_creative_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Creative ID.'
        },
        {
          name: 'appx_googlecm360_creative_name',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Creative name.'
        },
        {
          name: 'appx_googlecm360_creative_type',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Creative type.'
        },
        {
          name: 'appx_googlecm360_creative_version',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Creative version.'
        },
        {
          name: 'appx_googlecm360_site_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Website ID.'
        },
        {
          name: 'appx_googlecm360_site_name',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Website name.'
        },
        {
          name: 'appx_googlecm360_browser',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Browser.'
        },
        {
          name: 'appx_googlecm360_os',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 OS.'
        },
        {
          name: 'appx_googlecm360_segment_value',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Segment value.'
        },
        {
          name: 'appx_googlecm360_rendering_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'CM360 Rendering ID.'
        }
      ]
    }
  ]
}

const shopify: AppManifest = {
  id: 'appx_shopify',
  name: 'Shopify',
  homepage: 'https://www.rimdian.com/',
  author: 'Rimdian',
  icon_url: 'https://eu.rimdian.com/images/apps/shopify.png',
  short_description: 'Sync Shopify customers & orders',
  description:
    'Import your existing Shopify customers & orders. Keep your data up to date with webhooks.',
  version: '1.0.0',
  ui_endpoint: 'https://nativeapps.rimdian.com',
  webhook_endpoint: 'https://nativeapps.rimdian.com/api/webhooks',
  tasks: [
    {
      id: 'appx_shopify_sync',
      name: 'Import your Shopify customers & orders',
      is_cron: false,
      on_multiple_exec: 'discard_new'
    }
  ]
}

const googleAdsEC: AppManifest = {
  id: 'appx_googleadsec',
  name: 'Google Ads Enhanced Conversions',
  homepage: 'https://www.rimdian.com/',
  author: 'Rimdian',
  icon_url: 'https://eu.rimdian.com/images/apps/google-ads.png',
  short_description: 'Improve your Google Ads measurement with Enhanced Conversions.',
  description:
    'Improve the accuracy of your Google Ads conversions by sending first-party customer data in a privacy-safe way.',
  version: '1.0.0',
  ui_endpoint: 'https://nativeapps.rimdian.com',
  webhook_endpoint: 'https://nativeapps.rimdian.com/api/webhooks',
  data_hooks: [
    {
      id: 'appx_googleadsec_hook',
      name: 'Google Ads Enhanced Conversions API',
      on: 'on_success',
      for: [
        {
          kind: 'order',
          action: 'create'
        }
      ]
    }
  ]
}

const gclid: AppManifest = {
  id: 'appx_gclid',
  name: 'Google Ads Clicks',
  homepage: 'https://www.rimdian.com/',
  author: 'Rimdian',
  icon_url: 'https://eu.rimdian.com/images/apps/google-ads.png',
  short_description: 'Import your Google Ads clicks to enrich your sessions.',
  description:
    'Web sessions coming from Google Ads only contain a "gclid" parameter. Importing the Google clicks is mandatory to retrieve their corresponding campaigns, ads, search terms etc...',
  version: '1.0.0',
  ui_endpoint: 'https://nativeapps.rimdian.com',
  webhook_endpoint: 'https://nativeapps.rimdian.com/api/webhooks',
  sql_queries: [
    {
      id: 'appx_gclid_sessions',
      type: 'select',
      name: 'Fetch sessions with GCIDs',
      description:
        'Retrieve sessions and users core infos matching GCLIDs. Deduplicated by gclid (utm_id).',
      query:
        "SELECT s.external_id as session_external_id, s.user_id as user_id, s.domain_id as session_domain_id, s.created_at as session_created_at, s.utm_id as session_utm_id, u.external_id as user_external_id, u.is_authenticated as user_is_authenticated, u.created_at as user_created_at, u.timezone as user_timezone, u.language as user_language, u.country as user_country FROM `session` as s JOIN `user` as u ON s.user_id = u.id WHERE s.utm_id IN (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) AND s.utm_id_from = 'gclid' GROUP BY s.utm_id;",
      test_args: [
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a',
        'a'
      ]
    }
  ],
  tasks: [
    {
      id: 'appx_gclid_import',
      name: 'Import Google Ads clicks',
      is_cron: true,
      on_multiple_exec: 'discard_new',
      minutes_interval: 600
    }
  ],
  app_tables: [
    {
      name: 'appx_gclid_click',
      description: 'Google Ads Clicks',
      shard_key: ['external_id'],
      unique_key: ['external_id'],
      sort_key: ['created_at'],
      joins: [
        {
          external_table: 'session',
          relationship: 'one_to_one',
          local_column: 'external_id',
          external_column: 'utm_id'
        }
      ],
      columns: [
        {
          name: 'campaign_external_id',
          type: 'varchar',
          size: 256,
          is_required: true,
          description: 'Campaign external ID'
        },
        {
          name: 'campaign_name',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Campaign name'
        },
        {
          name: 'ad_group_external_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'Ad group external ID'
        },
        {
          name: 'ad_group_name',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Ad group name'
        },
        {
          name: 'ad_group_type',
          type: 'varchar',
          size: 64,
          is_required: false,
          description: 'Ad group type'
        },
        {
          name: 'ad_external_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'Ad external ID'
        },
        {
          name: 'keyword_external_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'Keyword external ID'
        },
        {
          name: 'keyword_name',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'Keyword name'
        },
        {
          name: 'keyword_match_type',
          type: 'varchar',
          size: 64,
          is_required: false,
          description: 'Keyword match type'
        },
        {
          name: 'ad_network_type',
          type: 'varchar',
          size: 64,
          is_required: false,
          description: 'Ad network type'
        },
        {
          name: 'ad_slot',
          type: 'varchar',
          size: 64,
          is_required: false,
          description: 'Ad slot'
        },
        {
          name: 'click_type',
          type: 'varchar',
          size: 64,
          is_required: false,
          description: 'Click type'
        },
        {
          name: 'user_list_external_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'User list (audience) external ID'
        },
        {
          name: 'id',
          type: 'varchar',
          size: 64,
          is_required: true,
          description: 'ID (sha1 of external_id)',
          hide_in_analytics: true
        },
        {
          name: 'external_id',
          type: 'varchar',
          size: 256,
          is_required: true,
          description: 'External ID',
          hide_in_analytics: true
        },
        {
          name: 'created_at',
          type: 'datetime',
          is_required: true,
          description: 'Created at'
        },
        {
          name: 'fields_timestamp',
          type: 'json',
          is_required: true,
          description: 'Fields timestamp',
          hide_in_analytics: true
        },
        {
          name: 'db_created_at',
          type: 'timestamp',
          size: 6,
          is_required: true,
          description: 'DB created at',
          default_timestamp: 'CURRENT_TIMESTAMP(6)',
          hide_in_analytics: true
        },
        {
          name: 'db_updated_at',
          type: 'timestamp',
          is_required: true,
          description: 'DB updated at',
          default_timestamp: 'CURRENT_TIMESTAMP',
          extra_definition: 'ON UPDATE CURRENT_TIMESTAMP',
          hide_in_analytics: true
        }
      ],
      timeseries_column: 'created_at',
      indexes: []
    }
  ]
}

const metaCapi: AppManifest = {
  id: 'appx_metacapi',
  name: 'Meta Conversions API',
  homepage: 'https://www.rimdian.com/',
  author: 'Rimdian',
  icon_url: 'https://eu.rimdian.com/images/apps/meta.png',
  short_description: 'Send your conversions to the Meta API.',
  description: 'Send your conversions to the Meta API.',
  version: '1.0.0',
  ui_endpoint: 'https://nativeapps.rimdian.com',
  webhook_endpoint: 'https://nativeapps.rimdian.com/api/webhooks',
  data_hooks: [
    {
      id: 'appx_metacapi_hook',
      name: 'Meta server-side conversions API',
      on: 'on_success',
      for: [
        {
          kind: 'order',
          action: 'create'
        },
        {
          kind: 'pageview',
          action: 'create'
        },
        {
          kind: 'cart_item',
          action: 'create'
        }
      ]
    }
  ]
}

const apps: AppManifest[] = [gclid, googleAdsEC, metaCapi, shopify, googleCM360]

export default apps
