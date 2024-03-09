export type Industry = {
  value: string
  label: string
}
const industries: Industry[] = [
  { value: 'arts-entertainment', label: 'Arts & Entertainment' },
  { value: 'automotive', label: 'Automotive' },
  { value: 'beauty-fitness', label: 'Beauty & Fitness' },
  { value: 'books-literature', label: 'Books & Literature' },
  { value: 'business-industrial-markets', label: 'Business & Industrial markets' },
  { value: 'computer-electronics', label: 'Computer & Electronics' },
  { value: 'finance', label: 'Finance' },
  { value: 'food-drink', label: 'Food & Drink' },
  { value: 'games', label: 'Games' },
  { value: 'healthcare', label: 'Healtcare' },
  { value: 'hobbies-leisure', label: 'Hobbies & Leisure' },
  { value: 'home-garden', label: 'Home & Garden' },
  { value: 'internet-telecom', label: 'Internet & Telecom' },
  { value: 'jobs-education', label: 'Jobs & Education' },
  { value: 'law-government', label: 'Law & Government' },
  { value: 'news', label: 'News' },
  { value: 'online-communities', label: 'Online communities' },
  { value: 'people-society', label: 'People & Society' },
  { value: 'pets-animals', label: 'Pets & Animals' },
  { value: 'real-estate', label: 'Real estate' },
  { value: 'science', label: 'Science' },
  { value: 'shopping', label: 'Shopping' },
  { value: 'sports', label: 'Sports' },
  { value: 'travel', label: 'Travel' },
  { value: 'other', label: 'Other' }
]

export default industries
