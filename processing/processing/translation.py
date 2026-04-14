"""Translation service using OpenRouter LLM."""

import structlog

import openai

from docvault_shared.config import config

logger = structlog.get_logger(__name__)


class TranslationResult:
    """Result from translation."""

    def __init__(
        self,
        original_text: str,
        translated_text: str,
        source_language: str,
        target_language: str,
        is_translation: bool,
    ):
        self.original_text = original_text
        self.translated_text = translated_text
        self.source_language = source_language
        self.target_language = target_language
        self.is_translation = is_translation
        self.chunk_index: int | None = None
        self.page_id: str | None = None


class OpenRouterTranslationService:
    """Translation service using OpenRouter LLM."""

    def __init__(self, api_key: str | None = None):
        """Initialize OpenRouter translation service."""
        self.api_key = api_key or config.openrouter_api_key
        self.model_name = config.translation_model.strip()
        if not self.model_name:
            raise ValueError("TRANSLATION_MODEL must not be empty")
        base_url = "https://openrouter.ai/api/v1"
        self.client = openai.OpenAI(api_key=self.api_key, base_url=base_url)

    async def translate_text(
        self,
        text: str,
        source_language: str,
        target_language: str = "en",
    ) -> TranslationResult:
        """Translate text to target language using OpenRouter LLM."""
        if target_language is None:
            target_language = "en"

        if source_language == target_language:
            return TranslationResult(
                original_text=text,
                translated_text=text,
                source_language=source_language,
                target_language=target_language,
                is_translation=False,
            )

        try:
            translated_text = await self._call_translation_api(
                text, source_language, target_language
            )

            logger.info(
                "text_translated",
                source_lang=source_language,
                target_lang=target_language,
                original_length=len(text),
                translated_length=len(translated_text),
            )

            return TranslationResult(
                original_text=text,
                translated_text=translated_text,
                source_language=source_language,
                target_language=target_language,
                is_translation=True,
            )

        except Exception as e:
            logger.error(
                "translation_failed",
                source_language=source_language,
                error=str(e),
            )
            raise

    async def _call_translation_api(
        self,
        text: str,
        source_language: str,
        target_language: str,
    ) -> str:
        """Call OpenRouter LLM for translation."""
        if not self.api_key:
            raise RuntimeError("OPENROUTER_API_KEY is required for translation")

        response = self.client.chat.completions.create(
            model=self.model_name,
            messages=[
                {
                    "role": "system",
                    "content": f"Translate the following text from {source_language} to {target_language}. "
                    f"Only respond with the translation, nothing else. Do not add quotes or explanations.",
                },
                {"role": "user", "content": text},
            ],
            max_tokens=4096,
        )

        translated = response.choices[0].message.content
        if not translated:
            raise Exception("No translation returned from OpenRouter")

        return translated.strip()

    async def translate_chunks(
        self,
        chunks: list,
        source_language: str,
    ) -> list[TranslationResult]:
        """Translate multiple text chunks."""
        results = []

        for chunk in chunks:
            result = await self.translate_text(
                chunk.text if hasattr(chunk, "text") else str(chunk),
                source_language,
            )
            if hasattr(chunk, "chunk_index"):
                result.chunk_index = chunk.chunk_index
            if hasattr(chunk, "page_id"):
                result.page_id = chunk.page_id
            results.append(result)

        logger.info(
            "chunks_translated",
            chunk_count=len(chunks),
            source_language=source_language,
        )

        return results


class DocumentTranslator:
    """Full document translation pipeline."""

    def __init__(self):
        """Initialize translator."""
        self.translation_service = OpenRouterTranslationService()

    async def translate_document(
        self,
        document_id: str,
        pages: list,
        source_language: str,
    ) -> dict:
        """Translate an entire document."""
        if source_language == "en":
            logger.info("document_english_skipped", document_id=document_id)
            return {
                "is_translation": False,
                "source_language": source_language,
                "pages": {},
            }

        logger.info(
            "translating_document",
            document_id=document_id,
            page_count=len(pages),
            source_language=source_language,
        )

        translated_pages = {}

        for page in pages:
            page_text = page.text if hasattr(page, "text") else str(page)
            page_number = page.page_number if hasattr(page, "page_number") else 0

            result = await self.translation_service.translate_text(
                page_text,
                source_language,
            )

            translated_pages[page_number] = {
                "original_text": result.original_text,
                "translated_text": result.translated_text,
                "is_translation": result.is_translation,
            }

        logger.info(
            "document_translated",
            document_id=document_id,
            translated_pages=len(translated_pages),
        )

        return {
            "is_translation": True,
            "source_language": source_language,
            "target_language": "en",
            "pages": translated_pages,
        }


translation_service = OpenRouterTranslationService()
document_translator = DocumentTranslator()
