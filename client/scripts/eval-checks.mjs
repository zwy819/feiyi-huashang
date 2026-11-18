export function checkAnswer(item, parsed, status) {
  const content = parsed.content || "";
  const citedIds = [...new Set([...content.matchAll(/\[S(\d+)\]/g)].map(match => `S${match[1]}`))];
  const availableIds = new Set(parsed.sources.map(source => source.id));
  const invalidCitations = citedIds.filter(id => !availableIds.has(id));
  const citedSources = parsed.sources.filter(source => citedIds.includes(source.id));
  const missingSources = (item.expected_sources || []).filter(title => !citedSources.some(source => source.title?.includes(title)));
  const missingSourceGroups = (item.expected_source_groups || []).filter(group => !group.some(title => citedSources.some(source => source.title?.includes(title))));
  const missingTerms = (item.expected_terms || []).filter(term => /^\d+$/.test(term)
    ? !new RegExp(`(?<![\\dA-Za-z])${term}(?![\\dA-Za-z])`).test(content.replace(/\[S\d+\]/g, ""))
    : !content.includes(term));
  const missingFacts = (item.expected_facts || []).filter(fact => !fact.any_of.some(term => content.includes(term)));
  const forbiddenTerms = (item.must_not_contain || []).filter(term => content.includes(term));
  const refusalOK = !item.should_refuse || /未找到直接依据|无法|不能|不提供|不应|拒绝|不足以|不便/.test(content);
  const citationOK = invalidCitations.length === 0 && (item.should_refuse || citedIds.length > 0);
  const hasCriteria = item.should_refuse || (item.expected_terms?.length || 0) + (item.expected_facts?.length || 0) > 0;
  return {
    automatic_ok: status === 200 && !!parsed.done && parsed.errors.length === 0 && content.trim().length > 0
      && citationOK && refusalOK && hasCriteria && missingSources.length === 0 && missingSourceGroups.length === 0
      && missingTerms.length === 0 && missingFacts.length === 0 && forbiddenTerms.length === 0,
    citation_ok: citationOK,
    refusal_ok: refusalOK,
    has_answer_criteria: hasCriteria,
    invalid_citations: invalidCitations,
    missing_sources: missingSources,
    missing_source_groups: missingSourceGroups,
    missing_terms: missingTerms,
    missing_facts: missingFacts,
    forbidden_terms: forbiddenTerms,
    cited_ids: citedIds,
  };
}
