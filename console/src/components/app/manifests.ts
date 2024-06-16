// disable no-template-curly-in-string
/* eslint-disable no-template-curly-in-string */

import { AppManifest, AppTable, CubeSchema } from 'interfaces'

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

const metaCapi: AppManifest = {
  id: 'appx_metacapi',
  name: 'Meta Conversions API',
  homepage: 'https://www.rimdian.com/',
  author: 'Rimdian',
  icon_url: 'https://eu.rimdian.com/images/apps/meta.png',
  short_description: 'Send your conversions to the Meta API.',
  description: 'Send your conversions to the Meta API.',
  version: '1.1.0',
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

const meta: AppManifest = {
  id: 'appx_meta',
  name: 'Meta',
  homepage: 'https://www.rimdian.com/',
  author: 'Rimdian',
  icon_url: 'https://eu.rimdian.com/images/apps/meta.png',
  short_description: 'Meta Conversions API & campaigns ROAS.',
  description:
    'Send your conversions to the Meta API and sync your campaigns stats to compute your ROAS.',
  version: '1.0.0',
  ui_endpoint: 'https://nativeapps.rimdian.com',
  webhook_endpoint: 'https://nativeapps.rimdian.com/api/webhooks',
  data_hooks: [
    {
      id: 'appx_meta_hook',
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
  ],
  tasks: [
    {
      id: 'appx_meta_import_metrics',
      name: 'Import Meta metrics',
      is_cron: true,
      on_multiple_exec: 'discard_new',
      minutes_interval: 720
    }
  ],
  app_tables: [
    // TODO
  ]
}

const googleAds: AppManifest = {
  id: 'appx_googleads',
  name: 'Google Ads',
  homepage: 'https://www.rimdian.com/',
  author: 'Rimdian',
  icon_url: 'https://eu.rimdian.com/images/apps/google-ads.png',
  short_description:
    'Import your Google Ads clicks & metrics (campaigns, ad groups, keywords...) to enrich your web sessions & compute your ROAS. Improve your Google Ads measurement with Enhanced Conversions.',
  description:
    'The Google Ads app automatically imports your ads clicks metadata (campaign, term, ad, cost...) to properly attribute the web sessions, and imports your campaigns, ad groups and keywords to analyze your ROAS. It also sends your conversions to the Google Ads API to improve the accuracy of your Google Ads conversions by sending first-party customer data in a privacy-safe way.',
  version: '2.0.0',
  ui_endpoint: 'https://nativeapps.rimdian.com',
  webhook_endpoint: 'https://nativeapps.rimdian.com/api/webhooks',
  sql_queries: [
    {
      id: 'appx_googleads_sessions',
      type: 'select',
      name: 'Fetch sessions with GCLIDs',
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
      id: 'appx_googleads_import_clicks',
      name: 'Import Google Ads clicks',
      is_cron: true,
      on_multiple_exec: 'discard_new',
      minutes_interval: 720
    },
    {
      id: 'appx_googleads_import_metrics',
      name: 'Import Google Ads metrics',
      is_cron: true,
      on_multiple_exec: 'discard_new',
      minutes_interval: 720
    }
  ],
  data_hooks: [
    {
      id: 'appx_googleads_enhanced_conversions',
      name: 'Google Ads Enhanced Conversions',
      on: 'on_success',
      for: [
        {
          kind: 'order',
          action: 'create'
        }
      ]
    }
  ],
  app_tables: [
    {
      name: 'appx_googleads_click',
      description: 'Google Ads clicks',
      shard_key: ['external_id'],
      unique_key: ['external_id'],
      sort_key: ['created_at'],
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
          name: 'ad_group_keyword_external_id',
          type: 'varchar',
          size: 256,
          is_required: false,
          description: 'Ad-group keyword external ID'
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
    },
    {
      name: 'appx_googleads_campaign',
      description: 'Google Ads campaign',
      shard_key: ['external_id'],
      unique_key: ['external_id'],
      sort_key: ['created_at'],
      columns: [
        {
          name: 'campaign_name',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Campaign name'
        },
        {
          name: 'resource_name',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Google Ads API resource name'
        },
        {
          name: 'start_date',
          type: 'varchar',
          size: 32,
          is_required: true,
          description: 'Campaign start date'
        },
        {
          name: 'status',
          type: 'varchar',
          size: 32,
          is_required: true,
          description: 'Campaign status'
        },
        {
          name: 'timezone',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Google Ads reporting timezone'
        },
        {
          name: 'advertising_channel_type',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Advertising channel type'
        },
        {
          name: 'advertising_channel_sub_type',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Advertising channel sub type'
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
    },
    {
      name: 'appx_googleads_campaign_metric',
      description: 'Google Ads campaigns daily metrics',
      shard_key: ['campaign_external_id', 'metrics_date'],
      unique_key: ['campaign_external_id', 'metrics_date'],
      sort_key: ['metrics_date'],
      indexes: [],
      columns: [
        {
          name: 'campaign_external_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Campaign external ID'
        },
        {
          name: 'metrics_date',
          type: 'varchar',
          size: 32,
          is_required: true,
          description: 'Date of the metrics'
        },
        {
          name: 'timezone',
          type: 'varchar',
          size: 32,
          is_required: true,
          description: 'Google Ads reporting timezone'
        },
        {
          name: 'clicks',
          type: 'number',
          is_required: true,
          description: 'Clicks'
        },
        {
          name: 'cost_micros',
          type: 'number',
          is_required: true,
          description: 'Cost (micros)'
        },
        {
          name: 'impressions',
          type: 'number',
          is_required: true,
          description: 'Impressions'
        },
        {
          name: 'conversions',
          type: 'number',
          is_required: true,
          description: 'Conversions reported by Google Ads'
        },
        {
          name: 'all_conversions',
          type: 'number',
          is_required: true,
          description:
            'The total number of conversions reported by Google Ads. This includes all conversions regardless of the value of include_in_conversions_metric.'
        },
        {
          name: 'absolute_top_impression_percentage',
          type: 'number',
          is_required: false,
          description:
            'The percent of your ad impressions that are shown as the very first ad above the organic search results.'
        },
        {
          name: 'active_view_cpm',
          type: 'number',
          is_required: false,
          description: 'Average cost of viewable impressions.'
        },
        {
          name: 'active_view_ctr',
          type: 'number',
          is_required: false,
          description:
            'Active view measurable clicks divided by active view viewable impressions. This metric is reported only for the Display Network.'
        },
        {
          name: 'active_view_impressions',
          type: 'number',
          is_required: false,
          description:
            'A measurement of how often your ad has become viewable on a Display Network site.'
        },
        {
          name: 'active_view_measurability',
          type: 'number',
          is_required: false,
          description:
            'The ratio of impressions that could be measured by Active View over the number of served impressions.'
        },
        {
          name: 'active_view_measurable_cost_micros',
          type: 'number',
          is_required: false,
          description:
            'The cost of the impressions you received that were measurable by Active View.'
        },
        {
          name: 'active_view_measurable_impressions',
          type: 'number',
          is_required: false,
          description:
            'The number of times your ads are appearing on placements in positions where they can be seen.'
        },
        {
          name: 'active_view_viewability',
          type: 'number',
          is_required: false,
          description:
            'The percentage of time when your ad appeared on an Active View enabled site (measurable impressions) and was viewable (viewable impressions).'
        },
        {
          name: 'all_conversions_from_click_to_call',
          type: 'number',
          is_required: false,
          description:
            'The number of times people clicked the "Call" button to call a store during or after clicking an ad. This number doesn\'t include whether or not calls were connected, or the duration of any calls. This metric applies to feed items only.'
        },
        {
          name: 'all_conversions_from_directions',
          type: 'number',
          is_required: false,
          description:
            'The number of times people clicked a "Get directions" button to navigate to a store after clicking an ad. This metric applies to feed items only.'
        },
        {
          name: 'all_conversions_from_interactions_rate',
          type: 'number',
          is_required: false,
          description:
            'All conversions from interactions (as oppose to view through conversions) divided by the number of ad interactions.'
        },
        {
          name: 'all_conversions_from_menu',
          type: 'number',
          is_required: false,
          description:
            "The number of times people clicked a link to view a store's menu after clicking an ad. This metric applies to feed items only."
        },
        {
          name: 'all_conversions_from_order',
          type: 'number',
          is_required: false,
          description:
            'The number of times people placed an order at a store after clicking an ad. This metric applies to feed items only.'
        },
        {
          name: 'all_conversions_from_other_engagement',
          type: 'number',
          is_required: false,
          description:
            'The number of other conversions (for example, posting a review or saving a location for a store) that occurred after people clicked an ad. This metric applies to feed items only.'
        },
        {
          name: 'all_conversions_from_store_visit',
          type: 'number',
          is_required: false,
          description:
            'Estimated number of times people visited a store after clicking an ad. This metric applies to feed items only.'
        },
        {
          name: 'all_conversions_from_store_website',
          type: 'number',
          is_required: false,
          description:
            "The number of times that people were taken to a store's URL after clicking an ad. This metric applies to feed items only."
        },
        {
          name: 'average_cost',
          type: 'number',
          is_required: false,
          description:
            'The average amount you pay per interaction. This amount is the total cost of your ads divided by the total number of interactions.'
        },
        {
          name: 'average_cpc',
          type: 'number',
          is_required: false,
          description:
            'The total cost of all clicks divided by the total number of clicks received.'
        },
        {
          name: 'average_cpe',
          type: 'number',
          is_required: false,
          description:
            "The average amount that you've been charged for an ad engagement. This amount is the total cost of all ad engagements divided by the total number of ad engagements."
        },
        {
          name: 'average_cpm',
          type: 'number',
          is_required: false,
          description: 'Average cost-per-thousand impressions (CPM).'
        },
        {
          name: 'average_cpv',
          type: 'number',
          is_required: false,
          description:
            'The average amount you pay each time someone views your ad. The average CPV is defined by the total cost of all ad views divided by the number of views.'
        },
        {
          name: 'conversions_from_interactions_rate',
          type: 'number',
          is_required: false,
          description:
            'Conversions from interactions divided by the number of ad interactions (such as clicks for text ads or views for video ads). This only includes conversion actions which include_in_conversions_metric attribute is set to true. If you use conversion-based bidding, your bid strategies will optimize for these conversions.'
        },
        {
          name: 'cost_per_all_conversions',
          type: 'number',
          is_required: false,
          description: 'The cost of ad interactions divided by all conversions.'
        },
        {
          name: 'cost_per_conversion',
          type: 'number',
          is_required: false,
          description:
            'The cost of ad interactions divided by conversions. This only includes conversion actions which include_in_conversions_metric attribute is set to true. If you use conversion-based bidding, your bid strategies will optimize for these conversions.'
        },
        {
          name: 'ctr',
          type: 'number',
          is_required: false,
          description:
            'The number of clicks your ad receives (Clicks) divided by the number of times your ad is shown (Impressions).'
        },
        {
          name: 'interaction_rate',
          type: 'number',
          is_required: false,
          description:
            'How often people interact with your ad after it is shown to them. This is the number of interactions divided by the number of times your ad is shown.'
        },
        {
          name: 'interactions',
          type: 'number',
          is_required: false,
          description:
            'The number of interactions. An interaction is the main user action associated with an ad format-clicks for text and shopping ads, views for video ads, and so on.'
        },
        {
          name: 'invalid_click_rate',
          type: 'number',
          is_required: false,
          description:
            'The percentage of clicks filtered out of your total number of clicks (filtered + non-filtered clicks) during the reporting period.'
        },
        {
          name: 'phone_calls',
          type: 'number',
          is_required: false,
          description: 'Number of offline phone calls.'
        },
        {
          name: 'phone_impressions',
          type: 'number',
          is_required: false,
          description: 'Number of offline phone impressions.'
        },
        {
          name: 'phone_through_rate',
          type: 'number',
          is_required: false,
          description:
            'Number of phone calls received (phone_calls) divided by the number of times your phone number is shown (phone_impressions).'
        },
        {
          name: 'relative_ctr',
          type: 'number',
          is_required: false,
          description:
            'Your clickthrough rate (Ctr) divided by the average clickthrough rate of all advertisers on the websites that show your ads. Measures how your ads perform on Display Network sites compared to other ads on the same sites.'
        },
        {
          name: 'search_absolute_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The percentage of the customer's Shopping or Search ad impressions that are shown in the most prominent Shopping position. See https://support.google.com/google-ads/answer/7501826 for details. Any value below 0.1 is reported as 0.0999."
        },
        {
          name: 'search_budget_lost_absolute_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The number estimating how often your ad wasn't the very first ad above the organic search results due to a low budget. Note: Search budget lost absolute top impression share is reported in the range of 0 to 0.9. Any value above 0.9 is reported as 0.9001."
        },
        {
          name: 'search_budget_lost_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The estimated percent of times that your ad was eligible to show on the Search Network but didn't because your budget was too low. Note: Search budget lost impression share is reported in the range of 0 to 0.9. Any value above 0.9 is reported as 0.9001."
        },
        {
          name: 'search_budget_lost_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The number estimating how often your ad didn't show anywhere above the organic search results due to a low budget. Note: Search budget lost top impression share is reported in the range of 0 to 0.9. Any value above 0.9 is reported as 0.9001."
        },
        {
          name: 'search_click_share',
          type: 'number',
          is_required: false,
          description:
            "The number of clicks you've received on the Search Network divided by the estimated number of clicks you were eligible to receive. Note: Search click share is reported in the range of 0.1 to 1. Any value below 0.1 is reported as 0.0999."
        },
        {
          name: 'search_exact_match_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The impressions you've received divided by the estimated number of impressions you were eligible to receive on the Search Network for search terms that matched your keywords exactly (or were close variants of your keyword), regardless of your keyword match types. Note: Search exact match impression share is reported in the range of 0.1 to 1. Any value below 0.1 is reported as 0.0999."
        },
        {
          name: 'search_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The impressions you've received on the Search Network divided by the estimated number of impressions you were eligible to receive. Note: Search impression share is reported in the range of 0.1 to 1. Any value below 0.1 is reported as 0.0999."
        },
        {
          name: 'search_rank_lost_absolute_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The number estimating how often your ad wasn't the very first ad above the organic search results due to poor Ad Rank. Note: Search rank lost absolute top impression share is reported in the range of 0 to 0.9. Any value above 0.9 is reported as 0.9001."
        },
        {
          name: 'search_rank_lost_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The estimated percentage of impressions on the Search Network that your ads didn't receive due to poor Ad Rank. Note: Search rank lost impression share is reported in the range of 0 to 0.9. Any value above 0.9 is reported as 0.9001."
        },
        {
          name: 'search_rank_lost_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The number estimating how often your ad didn't show anywhere above the organic search results due to poor Ad Rank. Note: Search rank lost top impression share is reported in the range of 0 to 0.9. Any value above 0.9 is reported as 0.9001."
        },
        {
          name: 'search_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The impressions you've received in the top location (anywhere above the organic search results) compared to the estimated number of impressions you were eligible to receive in the top location. Note: Search top impression share is reported in the range of 0.1 to 1. Any value below 0.1 is reported as 0.0999."
        },
        {
          name: 'top_impression_percentage',
          type: 'number',
          is_required: false,
          description:
            'The percent of your ad impressions that are shown anywhere above the organic search results.'
        },
        {
          name: 'video_quartile_p100_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched all of your video.'
        },
        {
          name: 'video_quartile_p25_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched 25% of your video.'
        },
        {
          name: 'video_quartile_p50_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched 50% of your video.'
        },
        {
          name: 'video_quartile_p75_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched 75% of your video.'
        },
        {
          name: 'video_view_rate',
          type: 'number',
          is_required: false,
          description:
            'The number of views your TrueView video ad receives divided by its number of impressions, including thumbnail impressions for TrueView in-display ads.'
        },
        {
          name: 'video_views',
          type: 'number',
          is_required: false,
          description: 'The number of times your video ads were viewed.'
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
      ]
    },
    {
      name: 'appx_googleads_ad_group_ad',
      description: 'Google Ads ad-group ads',
      shard_key: ['external_id'],
      unique_key: ['external_id'],
      sort_key: ['created_at'],
      timeseries_column: 'created_at',
      indexes: [],
      columns: [
        {
          name: 'resource_name',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Google Ads API ad-group ad resource name'
        },
        {
          name: 'status',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'The status of the ad.'
        },
        {
          name: 'ad_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Ad ID'
        },
        {
          name: 'ad_strength',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Overall ad strength for this ad group ad.'
        },
        {
          name: 'ad_name',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Ad name'
        },
        {
          name: 'ad_resource_name',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Google Ads API ad resource name'
        },
        {
          name: 'ad_type',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Ad type'
        },
        {
          name: 'ad_group_type',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Ad-group type'
        },
        {
          name: 'ad_group_external_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Ad-group external ID'
        },
        {
          name: 'ad_group_name',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Ad-group name'
        },
        {
          name: 'campaign_external_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Campaign external ID'
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
      ]
    },
    {
      name: 'appx_googleads_ad_group_ad_metric',
      description: 'Google Ads ad-group ads daily metrics',
      shard_key: ['campaign_external_id', 'metrics_date'],
      unique_key: ['campaign_external_id', 'ad_group_ad_external_id', 'metrics_date'],
      sort_key: ['metrics_date'],
      indexes: [],
      columns: [
        {
          name: 'campaign_external_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Campaign external ID'
        },
        {
          name: 'ad_group_ad_external_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Ad-group ad external ID'
        },
        {
          name: 'metrics_date',
          type: 'varchar',
          size: 32,
          is_required: true,
          description: 'Date of the metrics'
        },
        {
          name: 'timezone',
          type: 'varchar',
          size: 32,
          is_required: true,
          description: 'Google Ads reporting timezone'
        },
        {
          name: 'clicks',
          type: 'number',
          is_required: true,
          description: 'Clicks'
        },
        {
          name: 'cost_micros',
          type: 'number',
          is_required: true,
          description: 'Cost (micros)'
        },
        {
          name: 'impressions',
          type: 'number',
          is_required: true,
          description: 'Impressions'
        },
        {
          name: 'conversions',
          type: 'number',
          is_required: true,
          description: 'Conversions reported by Google Ads'
        },
        {
          name: 'all_conversions',
          type: 'number',
          is_required: true,
          description:
            'The total number of conversions reported by Google Ads. This includes all conversions regardless of the value of include_in_conversions_metric.'
        },
        {
          name: 'absolute_top_impression_percentage',
          type: 'number',
          is_required: false,
          description:
            'The percent of your ad impressions that are shown as the very first ad above the organic search results.'
        },
        {
          name: 'active_view_cpm',
          type: 'number',
          is_required: false,
          description: 'Average cost of viewable impressions.'
        },
        {
          name: 'active_view_ctr',
          type: 'number',
          is_required: false,
          description:
            'Active view measurable clicks divided by active view viewable impressions. This metric is reported only for the Display Network.'
        },
        {
          name: 'active_view_impressions',
          type: 'number',
          is_required: false,
          description:
            'A measurement of how often your ad has become viewable on a Display Network site.'
        },
        {
          name: 'active_view_measurability',
          type: 'number',
          is_required: false,
          description:
            'The ratio of impressions that could be measured by Active View over the number of served impressions.'
        },
        {
          name: 'active_view_measurable_cost_micros',
          type: 'number',
          is_required: false,
          description:
            'The cost of the impressions you received that were measurable by Active View.'
        },
        {
          name: 'active_view_measurable_impressions',
          type: 'number',
          is_required: false,
          description:
            'The number of times your ads are appearing on placements in positions where they can be seen.'
        },
        {
          name: 'active_view_viewability',
          type: 'number',
          is_required: false,
          description:
            'The percentage of time when your ad appeared on an Active View enabled site (measurable impressions) and was viewable (viewable impressions).'
        },
        {
          name: 'all_conversions_from_interactions_rate',
          type: 'number',
          is_required: false,
          description:
            'All conversions from interactions (as oppose to view through conversions) divided by the number of ad interactions.'
        },
        {
          name: 'average_cost',
          type: 'number',
          is_required: false,
          description:
            'The average amount you pay per interaction. This amount is the total cost of your ads divided by the total number of interactions.'
        },
        {
          name: 'average_cpc',
          type: 'number',
          is_required: false,
          description:
            'The total cost of all clicks divided by the total number of clicks received.'
        },
        {
          name: 'average_cpe',
          type: 'number',
          is_required: false,
          description:
            "The average amount that you've been charged for an ad engagement. This amount is the total cost of all ad engagements divided by the total number of ad engagements."
        },
        {
          name: 'average_cpm',
          type: 'number',
          is_required: false,
          description: 'Average cost-per-thousand impressions (CPM).'
        },
        {
          name: 'average_cpv',
          type: 'number',
          is_required: false,
          description:
            'The average amount you pay each time someone views your ad. The average CPV is defined by the total cost of all ad views divided by the number of views.'
        },
        {
          name: 'conversions_from_interactions_rate',
          type: 'number',
          is_required: false,
          description:
            'Conversions from interactions divided by the number of ad interactions (such as clicks for text ads or views for video ads). This only includes conversion actions which include_in_conversions_metric attribute is set to true. If you use conversion-based bidding, your bid strategies will optimize for these conversions.'
        },
        {
          name: 'cost_per_all_conversions',
          type: 'number',
          is_required: false,
          description: 'The cost of ad interactions divided by all conversions.'
        },
        {
          name: 'cost_per_conversion',
          type: 'number',
          is_required: false,
          description:
            'The cost of ad interactions divided by conversions. This only includes conversion actions which include_in_conversions_metric attribute is set to true. If you use conversion-based bidding, your bid strategies will optimize for these conversions.'
        },
        {
          name: 'ctr',
          type: 'number',
          is_required: false,
          description:
            'The number of clicks your ad receives (Clicks) divided by the number of times your ad is shown (Impressions).'
        },
        {
          name: 'interaction_rate',
          type: 'number',
          is_required: false,
          description:
            'How often people interact with your ad after it is shown to them. This is the number of interactions divided by the number of times your ad is shown.'
        },
        {
          name: 'interactions',
          type: 'number',
          is_required: false,
          description:
            'The number of interactions. An interaction is the main user action associated with an ad format-clicks for text and shopping ads, views for video ads, and so on.'
        },
        {
          name: 'top_impression_percentage',
          type: 'number',
          is_required: false,
          description:
            'The percent of your ad impressions that are shown anywhere above the organic search results.'
        },
        {
          name: 'video_quartile_p100_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched all of your video.'
        },
        {
          name: 'video_quartile_p25_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched 25% of your video.'
        },
        {
          name: 'video_quartile_p50_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched 50% of your video.'
        },
        {
          name: 'video_quartile_p75_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched 75% of your video.'
        },
        {
          name: 'video_view_rate',
          type: 'number',
          is_required: false,
          description:
            'The number of views your TrueView video ad receives divided by its number of impressions, including thumbnail impressions for TrueView in-display ads.'
        },
        {
          name: 'video_views',
          type: 'number',
          is_required: false,
          description: 'The number of times your video ads were viewed.'
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
      ]
    },
    {
      name: 'appx_googleads_ad_group_keyword',
      description: 'Google Ads keywords view',
      shard_key: ['external_id'],
      unique_key: ['external_id'],
      sort_key: ['created_at'],
      timeseries_column: 'created_at',
      indexes: [],
      columns: [
        {
          name: 'resource_name',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Google Ads API keyword view resource name'
        },
        {
          name: 'ad_group_external_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Ad-group external ID'
        },
        {
          name: 'campaign_external_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Campaign external ID'
        },
        {
          name: 'display_name',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Ad-group criterion display name'
        },
        {
          name: 'keyword_match_type',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Ad-group criterion keyword match type'
        },
        {
          name: 'keyword_text',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Ad-group criterion keyword text'
        },
        {
          name: 'criterion_negative',
          type: 'boolean',
          is_required: false,
          description: 'Whether to target (false) or exclude (true) the criterion.'
        },
        {
          name: 'status',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Ad-group criterion status'
        },
        {
          name: 'quality_score',
          type: 'number',
          is_required: false,
          description:
            'Ad-group criterion quality score. This field may not be populated if Google does not have enough information to determine a value.'
        },
        {
          name: 'creative_quality_score',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'The performance of the ad compared to other advertisers.'
        },
        {
          name: 'post_click_quality_score',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'The quality score of the landing page.'
        },
        {
          name: 'cpc_bid_micros',
          type: 'number',
          is_required: false,
          description: 'Ad-group criterion CPC bid micros'
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
      ]
    },
    {
      name: 'appx_googleads_ad_group_keyword_metric',
      description: 'Google Ads ad-group keywords daily metrics',
      shard_key: ['campaign_external_id', 'metrics_date'],
      unique_key: [
        'campaign_external_id',
        'ad_group_external_id',
        'ad_group_keyword_external_id',
        'metrics_date'
      ],
      sort_key: ['metrics_date'],
      indexes: [],
      columns: [
        {
          name: 'ad_group_external_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Ad-group external ID'
        },
        {
          name: 'campaign_external_id',
          type: 'varchar',
          size: 128,
          is_required: true,
          description: 'Campaign external ID'
        },
        {
          name: 'ad_group_keyword_external_id',
          type: 'varchar',
          size: 128,
          is_required: false,
          description: 'Ad-group criterion ID'
        },
        {
          name: 'clicks',
          type: 'number',
          is_required: false,
          description: 'Clicks'
        },
        {
          name: 'cost_micros',
          type: 'number',
          is_required: false,
          description: 'Cost (micros)'
        },
        {
          name: 'impressions',
          type: 'number',
          is_required: false,
          description: 'Impressions'
        },
        {
          name: 'conversions',
          type: 'number',
          is_required: false,
          description: 'Conversions reported by Google Ads'
        },
        {
          name: 'all_conversions',
          type: 'number',
          is_required: false,
          description:
            'The total number of conversions reported by Google Ads. This includes all conversions regardless of the value of include_in_conversions_metric.'
        },
        {
          name: 'active_view_cpm',
          type: 'number',
          is_required: false,
          description: 'Average cost of viewable impressions.'
        },
        {
          name: 'active_view_ctr',
          type: 'number',
          is_required: false,
          description:
            'Active view measurable clicks divided by active view viewable impressions. This metric is reported only for the Display Network.'
        },
        {
          name: 'active_view_impressions',
          type: 'number',
          is_required: false,
          description:
            'A measurement of how often your ad has become viewable on a Display Network site.'
        },
        {
          name: 'active_view_measurability',
          type: 'number',
          is_required: false,
          description:
            'The ratio of impressions that could be measured by Active View over the number of served impressions.'
        },
        {
          name: 'active_view_measurable_cost_micros',
          type: 'number',
          is_required: false,
          description:
            'The cost of the impressions you received that were measurable by Active View.'
        },
        {
          name: 'active_view_measurable_impressions',
          type: 'number',
          is_required: false,
          description:
            'The number of times your ads are appearing on placements in positions where they can be seen.'
        },
        {
          name: 'active_view_viewability',
          type: 'number',
          is_required: false,
          description:
            'The percentage of time when your ad appeared on an Active View enabled site (measurable impressions) and was viewable (viewable impressions)'
        },
        {
          name: 'all_conversions_from_interactions_rate',
          type: 'number',
          is_required: false,
          description:
            'All conversions from interactions (as oppose to view through conversions) divided by the number of ad interactions'
        },
        {
          name: 'average_cost',
          type: 'number',
          is_required: false,
          description:
            'The average amount you pay per interaction. This amount is the total cost of your ads divided by the total number of interactions'
        },
        {
          name: 'average_cpc',
          type: 'number',
          is_required: false,
          description: 'The total cost of all clicks divided by the total number of clicks received'
        },
        {
          name: 'average_cpe',
          type: 'number',
          is_required: false,
          description:
            "The average amount that you've been charged for an ad engagement. This amount is the total cost of all ad engagements divided by the total number of ad engagements"
        },
        {
          name: 'average_cpm',
          type: 'number',
          is_required: false,
          description: 'Average cost-per-thousand impressions (CPM)'
        },
        {
          name: 'average_cpv',
          type: 'number',
          is_required: false,
          description:
            'The average amount you pay each time someone views your ad. The average CPV is defined by the total cost of all ad views divided by the number of views'
        },
        {
          name: 'conversions_from_interactions_rate',
          type: 'number',
          is_required: false,
          description:
            'Conversions from interactions divided by the number of ad interactions (such as clicks for text ads or views for video ads). This only includes conversion actions which include_in_conversions_metric attribute is set to true. If you use conversion-based bidding, your bid strategies will optimize for these conversions'
        },
        {
          name: 'cost_per_all_conversions',
          type: 'number',
          is_required: false,
          description: 'The cost of ad interactions divided by all conversions'
        },
        {
          name: 'cost_per_conversion',
          type: 'number',
          is_required: false,
          description:
            'The cost of ad interactions divided by conversions. This only includes conversion actions which include_in_conversions_metric attribute is set to true. If you use conversion-based bidding, your bid strategies will optimize for these conversions'
        },
        {
          name: 'ctr',
          type: 'number',
          is_required: false,
          description:
            'The number of clicks your ad receives (Clicks) divided by the number of times your ad is shown (Impressions)'
        },
        {
          name: 'interaction_rate',
          type: 'number',
          is_required: false,
          description:
            'How often people interact with your ad after it is shown to them. This is the number of interactions divided by the number of times your ad is shown'
        },
        {
          name: 'interactions',
          type: 'number',
          is_required: false,
          description:
            'The number of interactions. An interaction is the main user action associated with an ad format-clicks for text and shopping ads, views for video ads, and so on'
        },
        {
          name: 'search_absolute_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The impressions you've received in the absolute top location (the very first ad above the organic search results) divided by the estimated number of impressions you were eligible to receive in the top location"
        },
        {
          name: 'search_budget_lost_absolute_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The estimated percent of times that your ad wasn't the very first ad above the organic search results due to a budget shortfall"
        },
        {
          name: 'search_budget_lost_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The estimated percent of times that your ad didn't show anywhere above the organic search results due to a budget shortfall"
        },
        {
          name: 'search_click_share',
          type: 'number',
          is_required: false,
          description:
            "The impressions you've received on the Search Network divided by the estimated number of impressions you were eligible to receive"
        },
        {
          name: 'search_exact_match_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The impressions you've received divided by the estimated number of impressions you were eligible to receive on the Search Network for search terms that matched your keywords exactly (or were close variants of your keyword)"
        },
        {
          name: 'search_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The impressions you've received on the Search Network divided by the estimated number of impressions you were eligible to receive"
        },
        {
          name: 'search_rank_lost_absolute_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The estimated percentage of impressions on the Search Network that your ads didn't receive due to poor Ad Rank in the absolute top (anywhere above the organic search results)"
        },
        {
          name: 'search_rank_lost_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The estimated percentage of impressions on the Search Network that your ads didn't receive due to poor Ad Rank"
        },
        {
          name: 'search_rank_lost_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The estimated percentage of impressions on the Search Network that your ads didn't receive due to poor Ad Rank"
        },
        {
          name: 'search_top_impression_share',
          type: 'number',
          is_required: false,
          description:
            "The impressions you've received in the top location (anywhere above the organic search results) compared to the estimated number of impressions you were eligible to receive in the top location"
        },
        {
          name: 'top_impression_percentage',
          type: 'number',
          is_required: false,
          description:
            'The percent of your ad impressions that are shown anywhere above the organic search results'
        },
        {
          name: 'video_quartile_p100_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched all of your video'
        },
        {
          name: 'video_quartile_p25_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched 25% of your video'
        },
        {
          name: 'video_quartile_p50_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched 50% of your video'
        },
        {
          name: 'video_quartile_p75_rate',
          type: 'number',
          is_required: false,
          description: 'Percentage of impressions where the viewer watched 75% of your video'
        },
        {
          name: 'video_view_rate',
          type: 'number',
          is_required: false,
          description:
            'The number of views your TrueView video ad receives divided by its number of impressions, including thumbnail impressions for TrueView in-display ads'
        },
        {
          name: 'video_views',
          type: 'number',
          is_required: false,
          description: 'The number of times your video ads were viewed'
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
          name: 'metrics_date',
          type: 'varchar',
          size: 32,
          is_required: true,
          description: 'Date of the metrics'
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
      ]
    }
  ],
  cube_schemas: {
    Appx_googleads_click: {
      sql: `SELECT * FROM appx_googleads_click`,
      title: 'Google Ads clicks',
      description: 'Google Ads clicks',
      joins: {
        Session: {
          relationship: 'one_to_one',
          sql: '${CUBE}.external_id = ${Session}.utm_id'
        }
      },
      measures: {
        count: {
          title: 'Count',
          type: 'count',
          sql: 'id',
          description: 'The number of clicks'
        }
      },
      dimensions: {}
    },
    Appx_googleads_campaign: {
      sql: `SELECT * FROM appx_googleads_campaign`,
      title: 'Google Ads campaigns',
      description: 'Google Ads campaigns',
      joins: {
        Session: {
          relationship: 'one_to_many',
          sql: '${CUBE}.campaign_name = ${Session}.utm_campaign'
        },
        Appx_googleads_campaign_metric: {
          relationship: 'one_to_many',
          sql: '${CUBE}.external_id = ${Appx_googleads_campaign_metric}.campaign_external_id'
        },
        Appx_googleads_ad_group_ad: {
          relationship: 'one_to_many',
          sql: '${CUBE}.external_id = ${Appx_googleads_ad_group_ad}.campaign_external_id'
        },
        Appx_googleads_ad_group_ad_metric: {
          relationship: 'one_to_many',
          sql: '${CUBE}.external_id = ${Appx_googleads_ad_group_ad_metric}.campaign_external_id'
        },
        Appx_googleads_ad_group_keyword: {
          relationship: 'one_to_many',
          sql: '${CUBE}.external_id = ${Appx_googleads_ad_group_keyword}.campaign_external_id'
        }
      },
      measures: {
        count: {
          title: 'Count',
          type: 'count',
          sql: 'id',
          description: 'The number of campaigns'
        }
      },
      dimensions: {}
    },
    Appx_googleads_campaign_metric: {
      title: 'Google Ads campaigns daily metrics',
      sql: `SELECT * FROM appx_googleads_campaign_metric`,
      description: 'Google Ads campaigns daily metrics',
      measures: {
        count: {
          title: 'Count',
          type: 'count',
          sql: 'id',
          description: 'The number of campaigns'
        }
      },
      dimensions: {}
    },
    Appx_googleads_ad_group_ad: {
      title: 'Google Ads ad group ads',
      sql: `SELECT * FROM appx_googleads_ad_group_ad`,
      description: 'Google Ads ad group ads',
      joins: {
        Session: {
          relationship: 'one_to_many',
          sql: '${CUBE}.resource_name = ${Session}.utm_content'
        },
        Appx_googleads_ad_group_ad_metric: {
          relationship: 'one_to_many',
          sql: '${CUBE}.external_id = ${Appx_googleads_ad_group_ad_metric}.ad_group_ad_external_id'
        }
      },
      measures: {
        count: {
          title: 'Count',
          type: 'count',
          sql: 'id',
          description: 'The number of ad group ads'
        }
      },
      dimensions: {}
    },
    Appx_googleads_ad_group_ad_metric: {
      title: 'Google Ads ad group ads daily metrics',
      sql: `SELECT * FROM appx_googleads_ad_group_ad_metric`,
      description: 'Google Ads ad group ads daily metrics',
      measures: {
        count: {
          title: 'Count',
          type: 'count',
          sql: 'id',
          description: 'The number of ad group ads'
        }
      },
      dimensions: {}
    },
    Appx_googleads_ad_group_keyword: {
      title: 'Google Ads keywords view',
      sql: `SELECT * FROM appx_googleads_ad_group_keyword`,
      description: 'Google Ads keywords view',
      joins: {
        Session: {
          relationship: 'one_to_many',
          sql: '${CUBE}.display_name = ${Session}.utm_term'
        },
        Appx_googleads_ad_group_keyword_metric: {
          relationship: 'one_to_many',
          sql: '${CUBE}.external_id = ${Appx_googleads_ad_group_keyword_metric}.ad_group_keyword_external_id'
        }
      },
      measures: {
        count: {
          title: 'Count',
          type: 'count',
          sql: 'id',
          description: 'The number of keywords'
        }
      },
      dimensions: {}
    },
    Appx_googleads_ad_group_keyword_metric: {
      title: 'Google Ads ad group keywords daily metrics',
      sql: `SELECT * FROM appx_googleads_ad_group_keyword_metric`,
      description: 'Google Ads ad group keywords daily metrics',
      measures: {
        count: {
          title: 'Count',
          type: 'count',
          sql: 'id',
          description: 'The number of keywords'
        }
      },
      dimensions: {}
    }
  }
}

// generate cube dimensions and measures (sum / avg)
const generateDimensionsAndMeasures = (table: AppTable, cubeSchema: CubeSchema) => {
  table.columns.forEach((column) => {
    if (column.type === 'number') {
      cubeSchema.measures = {
        ...cubeSchema.measures,
        [`${column.name}_sum`]: {
          title: `${column.name} sum`,
          type: 'sum',
          sql: column.name,
          description: `The sum of ${column.name}`
        },
        [`${column.name}_avg`]: {
          title: `${column.name} avg`,
          type: 'avg',
          sql: column.name,
          description: `The average of ${column.name}`
        }
      }
    }

    cubeSchema.dimensions[column.name] = {
      title: column.name,
      sql: column.name,
      type: 'string',
      description: column.description || column.name
    }

    if (column.name === 'id') {
      cubeSchema.dimensions[column.name].primaryKey = true
      cubeSchema.dimensions[column.name].shown = false
    }
  })
}

googleAds.app_tables?.forEach((table) => {
  if (!googleAds.cube_schemas) return
  const cubeName = table.name.charAt(0).toUpperCase() + table.name.slice(1)
  const cubeSchema = googleAds.cube_schemas[cubeName]
  if (!cubeSchema) return
  generateDimensionsAndMeasures(table, cubeSchema)
})

const wooCommerce: AppManifest = {
  id: 'appx_woocommerce',
  name: 'WooCommerce',
  homepage: 'https://www.rimdian.com/',
  author: 'Rimdian',
  icon_url: 'https://eu.rimdian.com/images/apps/woocommerce.png',
  short_description: 'Sync WooCommerce customers & orders',
  description:
    'Import your existing WooCommerce customers & orders. Keep your data up to date with webhooks.',
  version: '1.0.0',
  ui_endpoint: 'https://nativeapps.rimdian.com',
  webhook_endpoint: 'https://nativeapps.rimdian.com/api/webhooks',
  tasks: [
    {
      id: 'appx_woocommerce_sync',
      name: 'Import your WooCommerce customers & orders',
      is_cron: false,
      on_multiple_exec: 'discard_new'
    }
  ]
}

const apps: AppManifest[] = [googleAds, metaCapi, shopify, wooCommerce, meta, googleCM360]

export default apps
