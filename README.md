# CREC - A content recommendation and aggregation service

This service aggregates content from configurable providers and makes it accessible in a uniform JSON format. It supports various source formats (RSS, ATOM, JSON) and exposes API endpoints for retrieving content recommendations based on providers, topics/tags, full-text queries and locales.

It can also be used as an alternative Top Stories provider for the Recommended by section in Activity Stream (Your new tab page in Firefox), see [Activity Stream](#activity-stream) for details.

## Installing from Source

```
cd $GOPATH/src
go get github.com/csadilek/crec
cd github.com/csadilek/crec
go build
./crec
```

Steps to install Go: https://golang.org/doc/install

## Configuration

The service will start up on port 8080 using the default configuration, but customization is possible in ```config.toml```. Start by copying from ```config-example.toml``` which includes a set of all possible configuration options with explanations like the HTTP port setting.

```
cp config-example.toml config.toml
```

## Provider Registry

Content providers can be configured as .toml files in the specified provider registry dir (see config.toml). Here's an example of the New York Times feed for Space and Technology.

```
ID = "nyt-space"
Description = "NYT Space and Technology"
URL = "www.nytimes.com"
ContentURL = "http://rss.nytimes.com/services/xml/rss/nyt/Space.xml"
Categories = ["Space", "Technology"]
Processors = ["ExternalLinkRemover"]
Language = "en"
```

Only ```ID``` and ```ContentURL``` are mandatory. ```Categories``` can be used to specify defaults in case no categories are provided as part of the content. A list of content ```Processors``` can optionally be specified to modify content before ingestion.

## API

### Retrieve tag-based recommendations
```endpoint?t=[t1,t2]``` (t1 and t2 disjunctive) e.g. endpoint?t=Space,Technology (Space or Technology classified recommendations)

```endpoint?t=[t1]+[t2]``` (t1 and t2 conjunctive) e.g. endpoint?t=Sports+Triathlon (Sports and Triathlon classified recommendations)

### Retrieve query-based recommendations
```endpoint?q=[query]``` (searches the system’s full-text index for matching content)

### Retrieve provider-based recommendations
Provider based recommendations
```endpoint?p=[providerId]``` (returns content from the given provider)

### Retrieve locale-based recommendations
```endpoint?l=[locale]``` (returns content matching the given locale e.g. de-AT)

### Content push support
Providers can push content directly using a POST request to ```[endpoint]/crec/import``` using the system's content format. An API key has to be provided in the HTTP request’s Authorization header e.g. ```Authorization: APIKEY [content-provider-api-key]```.

API keys can be generated for all configured providers using ```crec -apiKeys```.

### Response format

The systems uniform response format looks as follows:

```
{
  "recommendations": [{
    "id": "http://www.nytimes.com/2017/03/10/science/space-dust-on-earth.html",
    "source": "nyt-space",
    "title": "Testing provider push mechanism 2- Flecks of Extraterrestrial Dust, All Over the Roof",
    "url": "http://www.nytimes.com/2017/03/10/science/space-dust.html?partner=rss\u0026emc=rss",
    "image_src": "https://static01.nyt.com/images/2017/03/11/science/14SCI-STARDUST-COMP01-moth.jpg",
    "explanation": "Selected for users interested in Space and Astronomy,Urban Areas,Meteors and Meteorites,Norway,Books and Literature,Space,Technology,Push",
    "author": "WILLIAM J. BROAD",
    "tags": ["Space and Astronomy", "Urban Areas", "Meteors and Meteorites", "Norway", "Books and Literature", "Space", "Technology", "Push"]
  },
  {
    "id": "https://www.nytimes.com/2017/09/17/science/occultation-moon-mars-venus-mercury.html",
    "source": "nyt-space",
    "title": "Trilobites: Three Planets Will Slide Behind the Moon in an Occultation",
    "url": "https://www.nytimes.com/2017/09/17/science/occultation-moon-mars-venus-mercury.html?partner=rss\u0026emc=rss",
    "image_src": "https://static01.nyt.com/images/2017/09/20/science/OCCULTATION/OCCULTATION-moth.jpg",
    "excerpt": "The moon will momentarily block Venus, then Mars and then Mercury, offering a vivid reminder of the cosmic clockwork of our solar system.",
    "explanation": "Selected for users interested in Moon,Mercury (Planet),Mars (Planet),Venus (Planet),Space and Astronomy,Space,Technology",
    "author": "NICHOLAS ST. FLEUR",
    "published_timestamp": "Sun, 17 Sep 2017 13:53:05 GMT",
    "tags": ["Moon", "Mercury (Planet)", "Mars (Planet)", "Venus (Planet)", "Space and Astronomy", "Space", "Technology"],
    "type": "recommended"
  }]
}
```

## Activity Stream

To use your deployed endpoint as a top story provider in activity stream:

- Browse to ```about:config```
- Click ```I accept the risk!```
- Type topstories.options in the Search field
- Double click ```browser.newtabpge.activity-stream.feeds.sections.topstories.options```
- Change the value to e.g.: ```{"provider_description": "My content selection", "provider_name": "Christian", "stories_endpoint": "http://[your-endpoint]/crec/content?t=Space+Technology"}```

