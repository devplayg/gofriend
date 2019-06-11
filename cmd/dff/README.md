## command 

### 중복파일 찾기

    ./dff -d /path/to/directory
    
### 다중 디렉토리에서 중복파일 찾기

    ./dff -d /path/to/directory -d /path/to/other/directory
    
### 최소 10개 이상 중복되는 파일 찾기

    ./dff -d /path/to/directory -c 10
    
### 최소 파일 크기가 2MB(2,000,000)이고, 최소 10개 이상 중복되는 파일 찾기

    ./dff -d /path/to/directory -s 2000000 -c 10 

### 결과 정렬

#### 파일그룹 전체 크기로 내림정렬

    ./dff -d /path/to/directory -r total

#### 단위 파일 크키로 내림 정렬
 
    ./dff -d /path/to/directory -r size
    
#### 파일 개수로 내림정렬    
    
    ./dff -d /path/to/directory -r count
    